bad-server
==========

bad-server is an http server that can be told to respond in various bad ways.
It is primarily intended for testing clients to make sure they're resilient
to server outages.

Usage
-----
bad-server figures out what behavior to exhibit based on headers you send it.

X-Response-Code-Histogram: send a status code based on an input histogram

  * X-Response-Code-Histogram: 500 => 100% 500 errors
  * X-Response-Code-Histogram: 500=50,200=50 will return 500s half the time, 200s the other
  * X-Response-Code-Histogram: 490 will return 490 even though it's not a standard http code

