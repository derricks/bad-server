bad-server
==========

bad-server is an http server that can be told to respond in various bad ways.
It is primarily intended for testing clients to make sure they're resilient
to server outages.

Usage
-----
bad-server figures out what behavior to exhibit based on headers you send it. In
general, the first header will win if you have multiple instances of the same
header in a request. Exceptions to this include:

  * X-Response-Code-Histogram
  * X-Return-Header
  
Headers
-------
X-Response-Code-Histogram: send a status code based on an input histogram

  * X-Response-Code-Histogram: 500 => 100% 500 errors
  * X-Response-Code-Histogram: 500=50,200=50 will return 500s half the time, 200s the other
  * X-Response-Code-Histogram: 490 will return 490 even though it's not a standard http code

X-Request-Body-As-Response: send the request body back in the response

X-Pause-Before-Response-Start: wait the specified amount of time before sending a response.

  * X-Pause-Before-Response-Start: 300 => wait 300ms before responding
  * X-Pause-Before-Response-Start: 1m => wait one minute before responding (uses [golang duration syntax](https://golang.org/pkg/time/#ParseDuration))
  
X-Add-Noise: randomly mutate bytes based on randomness percentage

  * X-Add-Noise: 3.0 => up to 3% of bytes will be randomly mutated
  
X-Return-Header: set the given header to the given value. Multiple instances of this header will get collated

  * X-Return-Header: Content-Type: application/json => response will have a Content-Type: application/json header
  
X-Generate-Random: generate random data for the given number of bytes

  * X-Generate-Random: 100 => 100 random bytes are returned
  
  
  