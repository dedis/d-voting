const express = require('express');
const path = require('path');
const axios = require('axios');

const app = express();

// Serve the static files from the React app
app.use(express.static(path.join(__dirname, 'client/build')));

// An api endpoint that returns a short list of items
app.get('/api/getTkKey', (req,res) => {

    body = 'urlaccess=http://128.179.185.30:3000/api/control_key\nservice=Evoting\nrequest=name,firstname,email,uniqueid,allunits';
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

                res.cookie('sign', sciper, { maxAge: 900000, httpOnly: true });
                res.redirect('/');

            } else {
                res.json('c est pas bon');
            }

        }).catch(error => {
            console.log(error);
    });

});




// Handles any requests that don't match the ones above
app.get('*', (req,res) =>{
    res.sendFile(path.join(__dirname+'/client/build/index.html'));
});

const port = process.env.PORT || 5000;
app.listen(port);

console.log('App is listening on port ' + port);