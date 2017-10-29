//gotp = require('gotp');

//var code = gotp('K2BBXLBE7PJDIM4TK2BBXLBE7PJDIM4T');
var code = "478298";
console.log('Try to check code', code);

var metadata = createMetadata({ apikey: 'secret_api_key' });
var call = client.check({
    type: 0,
    login: 'dev',
    code: code
}, metadata, pr);