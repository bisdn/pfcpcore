# PFCPCORE

_pfcpcore_ is a self contained library to support 3gpp PFCP based applications written in go language.
The only dependency outside go stdlib is the (easily replaced) logger.

_pfcpcore_ is not a _complete_ implementation of any specific 3gpp document, but is easily extensible to accommodate new IEs, etc., as required.
Its interoperability is most tested in R16 context.

_pfcpcore_ has some interesting, possibly unique features:

 - it is a validating parser; the validation is done in an automatically derived way, for every supported PFCP message type and group IE
 - the library provides rather a high level of support for client applications, so that,for example, specific IEs can be very easily located, based on ID,type code or customer predicate
 - the library incorporates an implementation of the PFCP 'reliable transport' mechanism
 - the library provides a generic (thus, flexible) way to merge the content of session establishment and session modification requests, which is an essential function in many real-world use cases
 - the library provides an implementation template/model for finite state machines to operate at the levels of association and session.  A typical user plane client application must implement such FSMs, and the library enables a very rapid implementation of an application by simply providing the customised call backs for each transition in the relevant state machine.

 The library is designed for use _as a library_, but is  bundled with a number of self contained applications, for example test endpoints which fulfil the roles of UPF or SMF.  One application in particular is noteworthy: 'sgwpgw'.  _sgwpgw_ performs a 4g-to-5g interworking function,such that a 4g core which is architected to use SGWu and PGWu functional blocks can interoperate with a 5g style aggregated user plane, i.e. UPF.

 # Missing Features

 _pfcpcore_ has no builtin support for charging, QoS, IPv6, ..... but, the framework can easily accommodate these functions, because its role is transparent to signaling message semantics.  The provided simple applications are for reference and potentially test applications.  _pfcpcore_ is designed to be extensible to support all current and future IEs, including for example fixed-line TR459 applications.  The framework for code generation to this goal is present in the repo, but is not yet enabled/completed.  A further, future, goal is to adapt/extend the library to provide a REDIS (or similar) structure, in which PFCP endpoint operations can be partitioned from forwarding plane specific functions.

 # Context

The main motivation for _pfcpcore_ is use for implementation of larger userplane applications than the included samples.

# Repository Structure

PFCP message encoding and decoding, validation and session state merge are in the main sub-directory 'pfcp'.

Reliable transport, and the needed underlying UDP socket handling, are in 'transport'.

Directories 'endpoint' and 'session' provide the higher level abstractions which can be used to build client applications.

# Contact

technical: Nicholas Hart - nphart@bisdn.de

BISDN GmbH
Körnerstraße 7-10, Aufgang A
10785 Berlin, Germany

https://www.bisdn.de/
