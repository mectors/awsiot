# AWS IoT Automatic Device Setup and MQTT wrapper

A run daemon will be automatically started. However it will fail if you don't run awsiot.init first and create cloud credentials.

This daemon will listen on a local MQTT server on topic cloud/aws/out/# [where # is a wildcard for any other topics]. If a message is received then it will send it to the AWS cloud to the same topic, e.g. cloud/aws/out/my/data/stream.

The same is try for the reverse. If send from cloud/aws/in/# then it will be send to the local MQTT server.

To set up, run one time:
  sudo /snap/bin/awsiot.init <key> <secret> <region>

Afterwards restart the service via:
  sudo snap disable awsiot
  sudo snap enable awsiot

This will create a $SNAP_COMMON/awscerts directory. Create an AWS policy, thing and certificates for this device.

Your device will automatically get a random 8 character thing name. You can find in the $SNAP_COMMON/awscerts directory:
the certificate:               certificate.crt
the private key:               private.key
the public key:                public.key
the root CA:                   rootca.pem
the iot configuration:         awsiot.json
the aws credentials:           awskeys.json
