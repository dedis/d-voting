#!/bin/sh






if [ $1 -eq 3 ]
then
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node1 start --port 2001"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node2 start --port 2002"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node3 start --port 2003"
elif [ $1 -eq 5 ]
then
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node1 start --port 2001"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node2 start --port 2002"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node3 start --port 2003"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node4 start --port 2004"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node5 start --port 2005"
elif [ $1 -eq 13 ]
then
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node1 start --port 2001"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node2 start --port 2002"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node3 start --port 2003"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node4 start --port 2004"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node5 start --port 2005"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node6 start --port 2006"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node7 start --port 2007"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node8 start --port 2008"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node9 start --port 2009"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node10 start --port 2010"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node11 start --port 2011"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node12 start --port 2012"
  gnome-terminal --tab --title="test" --command="./memcoin --config /tmp/node13 start --port 2013"
else
  echo "give 3, 5 or 13 "
fi



