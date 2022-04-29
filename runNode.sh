#!/bin/bash
# Basic while loop

pk=adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3




from=0
fromb=1
to=$1
ne=0
while [ $fromb -le $to ]
do
if [ $from -le 9 ]
then
gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node$fromb start --postinstall --promaddr :910$from --proxyaddr :908$from --proxykey $pk --listen tcp://0.0.0.0:200$fromb --public //localhost:200$fromb"

else



gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node$fromb start --postinstall --promaddr :91$from --proxyaddr :909$ne --proxykey $pk --listen tcp://0.0.0.0:20$fromb --public //localhost:20$fromb"
((ne++))
fi
((fromb++))
((from++))
done

