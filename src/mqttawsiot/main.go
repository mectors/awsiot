package main

import (
	"fmt"
	"flag"
	"os"
	"os/signal"
	"crypto/tls"
	"crypto/x509"
  "io/ioutil"
	"encoding/json"
	"strconv"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var conns = flag.Int("conns", 10, "how many conns (0 means infinite)")
var host = flag.String("host", "localhost:1883", "hostname of broker")
var user = flag.String("user", "", "username")
var pass = flag.String("pass", "", "password")
var dump = flag.Bool("dump", false, "dump messages?")
var wait = flag.Int("wait", 10, "ms to wait between client connects")
var pace = flag.Int("pace", 60, "send a message on average once every pace seconds")

var topic = "discovery"
var outtopic = "cloud/aws/out/#"
var intopic = "cloud/aws/in/#"

var Caws MQTT.Client
var Clocal MQTT.Client

func NewTlsConfig() *tls.Config {
	// Import trusted certificates from CAfile.pem.
	// Alternatively, manually add CA certificates to
	// default openssl CA bundle.
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile(os.Getenv("SNAP_COMMON")+"/awscerts/rootca.pem")
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	} else {
		fmt.Println("Please make sure you run sudo /snap/bin/awsiot.init <access key> <secret key> <region> before running this command.")
		panic(err)
	}

	// Import client certificate/key pair
	cert, err := tls.LoadX509KeyPair(os.Getenv("SNAP_COMMON")+"/awscerts/certificate.crt", os.Getenv("SNAP_COMMON")+"/awscerts/private.key")
	if err != nil {
		fmt.Println("Please make sure you run sudo /snap/bin/awsiot.init <access key> <secret key> <region> before running this command.")
		panic(err)
	}

	// Just to print out the client certificate..
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		panic(err)
	}

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: true,
		// Certificates = list of certs client sends to server.
		Certificates: []tls.Certificate{cert},
	}
}

// Send from AWS to local
var aws MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	//fmt.Println("from aws to local:"+msg.Topic()+":"+string(msg.Payload()))
	Clocal.Publish(msg.Topic(), 0, false,	msg.Payload())
}

// Send from local to AWS
var local MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
  //fmt.Println("from local to aws:"+msg.Topic()+":"+string(msg.Payload()))
	Caws.Publish(msg.Topic(), 0, false,	msg.Payload())
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func main() {

	// Set up channel on which to send signal notifications.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)

	// Prepare AWS secure connection to MQTT
	tlsconfig := NewTlsConfig()
	awsiotjson, err := ioutil.ReadFile(os.Getenv("SNAP_COMMON")+"/awscerts/awsiot.json")
	if err != nil {
		fmt.Println("Please make sure you run sudo /snap/bin/awsiot.init <access key> <secret key> <region> before running this command.")
		panic(err)
	}
  var dat map[string]interface{}
	err = json.Unmarshal(awsiotjson, &dat)
	check(err)
	var connection = "ssl://"+dat["host"].(string)+":"+strconv.FormatFloat(dat["port"].(float64), 'f', -1, 64)
	fmt.Println("Connecting to:"+connection)
	opts := MQTT.NewClientOptions()
	opts.AddBroker(connection)
	opts.SetClientID(dat["clientID"].(string)).SetTLSConfig(tlsconfig)
	opts.SetDefaultPublishHandler(aws)

	// Start the AWS connection
	Caws = MQTT.NewClient(opts)
	if token := Caws.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("Connected to aws")

  defer Caws.Disconnect(250)

	go func() {
	  // start listening for cloud/aws/in/<anything>
		if token := Caws.Subscribe(intopic, 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
		fmt.Println("Waiting for messages on: "+intopic)
	}()

  // Prepare the local MQTT connection
	opts2 := MQTT.NewClientOptions().AddBroker("tcp://"+*host)
  opts2.SetClientID(dat["clientID"].(string))
  opts2.SetDefaultPublishHandler(local)

  //create and start a client using the above ClientOptions
  Clocal = MQTT.NewClient(opts2)
  if token := Clocal.Connect(); token.Wait() && token.Error() != nil {
    panic(token.Error())
  }
	fmt.Println("Connected locally")
  defer Clocal.Disconnect(250)


	go func() {
	  // start listening for cloud/aws/out/<anything>
	  if token := Clocal.Subscribe(outtopic, 0, nil); token.Wait() && token.Error() != nil {
	    fmt.Println(token.Error())
	    os.Exit(1)
	  }
		fmt.Println("Waiting for messages on: "+outtopic)
	}()

	// Say we are ready for action
	Clocal.Publish(topic, 0, false,	outtopic)
  fmt.Println("Published to:"+topic+" that we are listening on:"+outtopic)

	// loop while waiting for commands to come in
	// Wait for receiving a signal.
	<-sigc
}
