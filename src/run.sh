#!/bin/sh
args="$SNAP $SNAP_COMMON"
for arg in "$@";
do
  args="$args $arg"
done
$SNAP/bin/mqttawsiot $args
