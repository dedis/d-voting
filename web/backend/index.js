const express = require('express');
const path = require('path');
const axios = require('axios');
const cookieParser = require("cookie-parser");
const session = require('express-session');

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

    body = 'urlaccess=http://128.179.189.204:3000/api/control_key\nservice=Evoting\nrequest=name,firstname,email,uniqueid,allunits';
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

                req.session.userid = sciper;

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
    if(req.session.userid){

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