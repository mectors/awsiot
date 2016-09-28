//var awsIot = require('aws-iot-device-sdk');

// Write creatething.json - TODO make thingname configurable
var fs = require('fs');
var common   = process.env.SNAP_COMMON;
var snap = process.env.SNAP;
var randomstring = require("randomstring");
var thingname = randomstring.generate(8);
var awscerts = common + '/awscerts';
var awsenv = awscerts+'/awsenv.json';

if (process.argv.length <= 3) {
  console.log("Usage: " + __filename + " <KEY> <SECRET> <REGION>");
  process.exit(-1);
}
var key = process.argv[2];
var secret = process.argv[3];
var region = process.argv[4];

// Read and set AWS credentials
if (!fs.existsSync(awscerts)){
    fs.mkdirSync(awscerts);
}
process.env['AWS_ACCESS_KEY_ID'] = key;
process.env['AWS_SECRET_ACCESS_KEY'] = secret;
process.env['AWS_DEFAULT_REGION'] = region;


// Create a thing
var shell = require('shelljs');
var cmd = snap + "/bin/aws iot create-thing --thing-name "+thingname+" --attribute-payload attributes={creator=aws.snap} > "+common+"/ctresponse.json";
console.log(cmd);
var output = shell.exec(cmd, {silent:false}).output;

// Create a policy
var cmd = snap + "/bin/aws iot create-policy --policy-name "+thingname+"policy --policy-document file://policy.json > "+common+"/cpresponse.json";
console.log(cmd);
var output = shell.exec(cmd, {silent:false}).output;

// Create AWS certificates
var certsdir = common+'/awscerts';
if (!fs.existsSync(certsdir)){
    fs.mkdirSync(certsdir);
}
var cmd = snap + "/bin/aws iot create-keys-and-certificate --set-as-active > "+certsdir+"/certificates.json";
console.log(cmd);
var output = shell.exec(cmd, {silent:false}).output;
var certificates = require(certsdir+"/certificates.json");

// Attach certificate to thing
var cmd = snap + "/bin/aws iot attach-thing-principal --thing-name "+thingname+" --principal "+certificates.certificateArn;
console.log(cmd);
var output = shell.exec(cmd, {silent:false}).output;

// Attach certificate to policy
var cmd = snap + "/bin/aws iot attach-principal-policy --policy-name "+thingname+"policy --principal "+certificates.certificateArn + " > "+common+"/apresponse.json";
console.log(cmd);
var output = shell.exec(cmd, {silent:false}).output;

// Describe end point
var cmd = snap + "/bin/aws iot describe-endpoint > " + common + "/endpoint.json";
console.log(cmd);
var output = shell.exec(cmd, {silent:false}).output;
var endpoint = require(common + "/endpoint.json");

// Write out the AWS environment
var stream = fs.createWriteStream(certsdir + "/awskeys.json");
awskeysjson = "{\n\t\"accesskey\": \"" + key + "\",\n\t\"secretkey\": \"" + secret + "\",\n\t\"region\": \""+region+"\"\n}";
stream.write(awskeysjson);
stream.end();
console.log("wrote "+ certsdir + "/awskeys.json");

// Write out the AWS IoT setup
var stream = fs.createWriteStream(certsdir + "/awsiot.json");
awsiotjson = "{\n\t\"host\": \"" + endpoint.endpointAddress + "\",\n\t\"port\": 8883,\n\t\"clientID\": \"" + thingname + "\",\n\t\"thingName\": \"" + thingname + "\",\n\t\"caCert\": \"awscerts/rootca.pem\",\n\t\"clientCert\": \"awscerts/certificate.crt\",\n\t\"privateKey\": \"awscerts/private.key\"\n}";
stream.write(awsiotjson);
stream.end();
console.log("wrote "+ certsdir + "/awsiot.json");

// Write out the certificates
var stream = fs.createWriteStream(certsdir + "/certificate.crt");
stream.write(certificates.certificatePem);
stream.end();
console.log("wrote "+ certsdir + "/certificate.crt");

var stream = fs.createWriteStream(certsdir + "/public.key");
stream.write(certificates.keyPair.PublicKey);
stream.end();
console.log("wrote "+ certsdir + "/public.key");

var stream = fs.createWriteStream(certsdir + "/private.key");
stream.write(certificates.keyPair.PrivateKey);
stream.end();
console.log("wrote "+ certsdir + "/private.key");

var rootca = "conf/rootca.pem";
if (!fs.existsSync(certsdir + "/rootca.pem")){
   fs.createReadStream(rootca).pipe(fs.createWriteStream(certsdir + "/rootca.pem"));
   console.log("moved "+ certsdir + "/rootca.pem");
}
