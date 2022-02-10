const express = require('express');
const path = require('path');
const axios = require('axios');
const cookieParser = require("cookie-parser");
const session = require('express-session');
const kyber = require('@dedis/kyber')
const crypto = require('crypto');
const request = require('request');
const config = require('./config.json');
const access_config = require('./access_config.json');
const lmdb = require('lmdb');

/*global Buffer, __dirname, process */

const app = express();

// Serve the static files from the React app
app.use(express.static(path.join(__dirname, 'client/build')));

//Express-session
app.set('trust-proxy', 1);

app.use(cookieParser());
const oneDay = 1000 * 60 * 60 * 24;
app.use(session({
    secret: config.SESSION_SECRET,
    saveUninitialized:true,
    cookie: { maxAge: oneDay },
    resave: false
}));

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

var access_control = function (req, res, next){
    const begin = req.url.split('?')[0];
    var role = 'everyone';
    if(req.session.userid){
        role = req.session.role;
    }

    if(access_config[role].includes(begin)){
        next();
    } else {
        res.status(400).send('Unauthorized');
    }
}
app.use(access_control);

let usersDB = lmdb.open({path: 'dvoting-users'});

/*
* This is via this endpoint that the client request the tequila key, this key will then be used for redirection on the tequila server
* */
app.get('/api/getTkKey', (req,res) => {

    const body = 'urlaccess=' + config.FRONT_END_URL + '/api/control_key\nservice=Evoting\nrequest=name,firstname,email,uniqueid,allunits';
    axios 
        .post('http://tequila.epfl.ch/cgi-bin/tequila/createrequest', body)
        .then(response => {
            const key = response.data.split('\n')[0].split('=')[1];
            const url = 'https://tequila.epfl.ch/cgi-bin/tequila/requestauth?requestkey=' + key;
            res.json({'url' : url});
        })
        .catch(error => {
            console.error(error)
        });
});


/*
* Here the client will send the key he/she received from the tequila, it is then verified on the tequila.
* If the key is valid, the user is then logged in the website through this backend
*/
app.get('/api/control_key', (req, res) => {
    const usr_key = req.query.key;
    const body = 'key=' + usr_key;

    axios
        .post('https://tequila.epfl.ch/cgi-bin/tequila/fetchattributes', body)
        .then(resa => {
            if(resa.data.includes('status=ok')){
                const sciper = resa.data.split('uniqueid=')[1].split('\n')[0];
                const lastname = resa.data.split('\nname=')[1].split('\n')[0];
                const firstname = resa.data.split('\nfirstname=')[1].split('\n')[0];

                var user = usersDB.get(sciper) || {};
                if(user["role"] === undefined || user["role"] === "") {
                    user["role"] = "voter";
                    user["lastname"] = lastname;
                    user["firstname"] = firstname;
                    user["loggedin"] = false;
                }
                usersDB.put(sciper, user);

                req.session.userid = parseInt(sciper);
                req.session.lastname = lastname;
                req.session.firstname = firstname;
                req.session.role = user["role"];
                res.redirect('/');
            } else {
                res.status(500).send('Login did not work')
            }

        }).catch(error => {
            console.log(error);
    });
});

/*
*  This endpoint serves to logout from the app by clearing the session
*/
app.get('/api/logout', (req, res) => {
    req.session.destroy();
    res.redirect('/');
});

/*
 * As the user is logged on the app via this express but must also be logged in the react.
 * This endpoint serves to send to the client (actually to react) the information of the current user
 */
app.get('/api/getpersonnalinfo', (req, res) => {

    if(req.session.userid){
        res.json({
            'sciper' : req.session.userid,
            'lastname' : req.session.lastname,
            'firstname' : req.session.firstname,
            'role' : req.session.role,
            'islogged' : true
        });
    } else {
        res.json({
            'sciper' : 0,
            'lastname' : '',
            'firstname' : '',
            'role' : '',
            'islogged' : false
        });
    }
});


/*
* This call allow a user that is admin to get the list of the poeple that have a special role (not a voter)
*/
app.get('/api/get_user_rights', (req, res) => {
    var sciper = req.session.userid;

    if(sciper){
        var user = usersDB.get(sciper);
        if(user["role"] === "admin") {
            var users = [];
            usersDB.getRange().filter(({key, value}) => {
                    return key != sciper; // filter out self
                })
                .forEach(({key, value}) => {
                    users.push(value);
                });
        } else {
            res.status(400).send('You must be admin to request this');
        }

        res.json(users);
    } else {
        res.status(400).send('Not logged in');
    }
});


/*
* This call (only for admins) allow an admin to add a role to a voter
*/
app.post('/api/add_role', (req, res) => {
    if(req.session.userid){
        var requester = usersDB.get(req.session.userid);
        if(requester["role"] === "admin") {

            const sciper = req.body.sciper;
            const role = req.body.role;
            var user = usersDB.get(sciper);
            user["role"] = role;
            usersDB.put(sciper, user);

        } else {
            res.status(400).send('You must be admin to request this');
        }
    } else {
        res.status(400).send('Not logged in');
    }
});


/*
* This call (only for admins) allow an admin to remove a role to a user
*/
app.post('/api/remove_role', (req, res) => {
    if(req.session.userid){
        if(req.session.role == 'admin'){
            const sciper = req.body.sciper;

            usersDB.remove(sciper);
            res.status(200).send('Deleted');
        } else {
            res.status(400).send('You must be admin to request this')
        }
    } else {
        res.status(400).send('Not logged in');
    }
});


/*
* This API call is used redirect all the calls for DELA to the DELAs nodes.
* During this process the data are processed : the user is authenticated and controlled.
* Once this is done the data are signed before the are sent to the DELA node
* To make this work, react has to redirect to this backend all the request that needs to go the DELA nodes
*/
app.post('/evoting/*', (req, res) => {

    //check session
    if(req.session.userid){

        const body_data = req.body;
        body_data["AdminID"] = req.session.userid;
        body_data["UserID"] = req.session.userid;
        const data_str = JSON.stringify(body_data);
        const data_str_b64 = Buffer.from(data_str).toString('base64');

        const hash = crypto.createHash('sha256').update(data_str_b64).digest('base64');

        const edCurve = kyber.curve.newCurve("edwards25519");

        const priv = Buffer.from(config.PRIVATE_KEY, 'hex');
        const pub = Buffer.from(config.PUBLIC_KEY, 'hex');

        const scalar = edCurve.scalar();
        scalar.unmarshalBinary(priv);

        const point = edCurve.point();
        point.unmarshalBinary(pub);

        const sign = kyber.sign.schnorr.sign(edCurve, scalar, hash);

        const payload = {
            'payload' : data_str_b64,
            'sign' : sign
        }

        var clientServerOptions = {
            uri: config.DELA_NODE_URL + req.url,
            body: JSON.stringify(payload),
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        }
        request(clientServerOptions, function (error, response) {
            console.log(error);
            console.log(response);

            res.json(response.body);
        });
    } else {
        res.status(400).send('Unauthorized');
    }

});

// Handles any requests that don't match the ones above
app.get('*', (req,res) =>{
    res.sendFile(path.join(__dirname+'/client/build/index.html'));
});


const port = process.env.PORT || 5000;
app.listen(port);

console.log('App is listening on port ' + port);