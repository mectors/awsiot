# AWS IoT Automatic Device Setup

Run one time:
  sudo /snap/bin/awsiot.init <key> <secret> <region>

This will create a $SNAP_COMMON/awscerts directory. Create an AWS policy, thing and certificates for this device.

Your device will automatically get a random 8 character thing name. You can find in the $SNAP_COMMON/awscerts directory:
the certificate:               certificate.crt
the private key:               private.key
the public key:                public.key
the root CA:                   rootca.pem
the iot configuration:         awsiot.json
the aws credentials:           awskeys.json
