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
  
Complex Examples
----------------
Combining headers lets you create other types of unexpected behaviors.

Response's content-type does not match content.

    X-Return-Header: Content-Type: application/json
    X-Generate-Random: 600
    
Response's Content-Length is longer than content.

    X-Return-Header: Content-Length: 6000
    X-Generate-Random: 1000
    
Response's Content-Length is shorter than content.

    X-Return-Header: Content-Length: 1000
    X-Generate-Random: 6000
    
Development
-----------
bad-server works by creating an ordered pipeline of functions that all take http.ResponseWriter
pointers. Functions are placed into the pipeline in this order:

  1. status_code generator
  2. header generators
  3. response body generator wrapped in response affectors
  
Only one status code generator and response body generator will be derived
from the headers sent to bad-server. If you include multiple response body
generators in your header, they will be processed in this order (from `general.go`)

    1. X-Request-Body-As-Response
    2. X-Generate-Random
    3. empty string
  
  
  