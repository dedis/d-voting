const express = require('express');
const path = require('path');
const axios = require('axios');
const cookieParser = require("cookie-parser");
const session = require('express-session');
//import kyber from "@dedis/kyber";
const kyber = require('@dedis/kyber')
const crypto = require('crypto');
const fetch = require('node-fetch');
const request = require('request');


const test = kyber.curve.newCurve('edwards25519');

const app = express();

// Serve the static files from the React app
app.use(express.static(path.join(__dirname, 'client/build')));

//Express-session
app.set('trust-proxy', 1);

app.use(cookieParser());
const oneDay = 1000 * 60 * 60 * 24;
app.use(session({
    secret: "thisismysecrctekeyfhrgfgrfrty84fwir767",
    saveUninitialized:true,
    cookie: { maxAge: oneDay },
    resave: false
}));

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

//serving public file
//app.use(express.static(__dirname));


// An api endpoint that returns a short list of items
app.get('/api/getTkKey', (req,res) => {

    body = 'urlaccess=https://dvoting-dev.dedis.ch/api/control_key\nservice=Evoting\nrequest=name,firstname,email,uniqueid,allunits';
    axios
        .post('http://tequila.epfl.ch/cgi-bin/tequila/createrequest', body)
        .then(resa => {
            key = resa.data.split('\n')[0].split('=')[1];
            url = 'https://tequila.epfl.ch/cgi-bin/tequila/requestauth?requestkey=' + key;
            res.json({'url' : url});
        })
        .catch(error => {
            console.error(error)
        });
});

app.get('/api/control_key', (req, res) => {
    usr_key = req.query.key;
    body = 'key=' + usr_key;

    axios
        .post('https://tequila.epfl.ch/cgi-bin/tequila/fetchattributes', body)
        .then(resa => {
            if(resa.data.includes('status=ok')){
                //res.json(resa.data);
                sciper = resa.data.split('uniqueid=')[1].split('\n')[0];
                name = resa.data.split('name=')[1].split('\n')[0];
                firstname = resa.data.split('firstname=')[1].split('\n')[0];

                req.session.userid = parseInt(sciper);
                req.session.name = name;
                req.session.firstname = firstname;

                //res.cookie('sign', sciper, { maxAge: 900000, httpOnly: true }); //TODO change sciper by signed message and maxAge
                res.redirect('/');

            } else {
                res.json('c est pas bon');
            }

        }).catch(error => {
            console.log(error);
    });

});

app.get('/api/logout', (req, res) => {
    req.session.destroy();
    res.json('ok');
});

app.get('/api/getpersonnalinfo', (req, res) => {

    if(req.session.userid){
        res.json({
            'sciper' : req.session.userid,
            'name' : req.session.name,
            'firstname' : req.session.firstname,
            'role' : 'voter',
            'islogged' : true
        });
    } else {
        res.json({
            'sciper' : 0,
            'name' : '',
            'firstname' : '',
            'role' : '',
            'islogged' : false
        });
    }
});

app.get('/evoting/*', (req, res) => {



    //check session
    if(req.session.userid){
        url = req.originalUrl;
    } else {
        //not logged in
    }

    res.json({
       'request' : 'receveid',
       'status' : 'ok'
    });

});

app.post('/evoting/*', (req, res) => {

    //check session
    if(req.session.userid || true){ //TODO change this condition
        const data_str = JSON.stringify(req.body);
        const data_str_b64 = Buffer.from(data_str).toString('base64');

        const hash = crypto.createHash('sha256').update(data_str_b64).digest('base64');

        const edCurve = kyber.curve.newCurve("edwards25519");

        const priv = Buffer.from('0435895685b8265285d411e37572567b6d7e6aef2a06e809d5ccd26ee050cc07', 'hex'); //TODO put that in separate file
        const pub = Buffer.from('ee4eb5309f46c02c57d78cd60fa4b7c6875626b0b17078c70f24930fcc8c5e62', 'hex'); //TODO put that in separate file

        const scalar = edCurve.scalar();
        scalar.unmarshalBinary(priv);

        const point = edCurve.point();
        point.unmarshalBinary(pub);

        const sign = kyber.sign.schnorr.sign(edCurve, scalar, hash);

        payload = {
            'payload' : data_str_b64,
            'sign' : sign
        }


        fetch('http://localhost:1000' + req.url) //TODO ajouter un truc pour sanitize l url
            .then(a => a.json())
            .then((result) => {

                console.log(result);
            });

        res.json(payload);

    } else {
        //not logged in
    }

    res.json({
        'request' : 'receveid',
        'status' : 'ok'
    });

});


// Handles any requests that don't match the ones above
app.get('*', (req,res) =>{
    res.sendFile(path.join(__dirname+'/client/build/index.html'));
});

const port = process.env.PORT || 5000;
app.listen(port);

console.log('App is listening on port ' + port);