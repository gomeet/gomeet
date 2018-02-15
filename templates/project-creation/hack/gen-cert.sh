#!/bin/bash

if [ "$1" = "" ]
then
    echo "usage: $0 service_name ca_name"
    exit 1
fi

if [ "$2" = "" ]
then
    echo "usage: $0 service_name ca_name"
    exit 1
fi

CERTNAME=$1
CANAME=$2

# Define base directory
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do # resolve $SOURCE until the file is no longer a symlink
  TARGET="$(readlink "$SOURCE")"
  if [[ $TARGET == /* ]]; then
    echo "SOURCE '$SOURCE' is an absolute symlink to '$TARGET'"
    SOURCE="$TARGET"
  else
    DIR="$( dirname "$SOURCE" )"
    echo "SOURCE '$SOURCE' is a relative symlink to '$TARGET' (relative to '$DIR')"
    SOURCE="$DIR/$TARGET" # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
  fi
done
RDIR="$( dirname "$SOURCE" )"
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

cd $DIR/../


if [ ! -d "data/certs" ]
then
    echo "Directory data/certs does not exist"
    exit 1
fi

if ! [ -x "$(command -v openssl)" ]; then
    echo "The openssl command is unavailable"
    exit 1
fi

if ! [ -x "$(command -v _tools/bin/retool)" ] || ! [ -x "$(command -v _tools/bin/certstrap)" ]; then
    echo "The retool command is unavailable."
    make tools
    if ! [ -x "$(command -v _tools/bin/retool)" ] || ! [ -x "$(command -v _tools/bin/certstrap)" ]; then
        echo "The retool or certstrap command is unavailable and make tools doesn't install it"
        echo "Check Makefile"
        exit 1
    fi
fi

openssl genrsa -out $CERTNAME.key 2048
openssl req -new -sha256 -key  $CERTNAME.key -out  $CERTNAME.csr -days 3650
mv $CERTNAME.key $CERTNAME.csr data/certs/
echo
echo "Invoking certstrap to sign the certificate request..."
_tools/bin/retool do certstrap --depot-path "data/certs" sign --CA $CANAME $CERTNAME
echo
echo "Converting keypair in PKCS#12 format for Android (use an empty export password)..."
openssl pkcs12 -export -inkey data/certs/$CERTNAME.key -in data/certs/$CERTNAME.crt -out data/certs/$CERTNAME.p12
exit 0
