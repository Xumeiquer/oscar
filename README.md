# OSCAR
Oscar is just a simple name for a simple DNS resolver with HTTP API capabilities.

Oscar implements an authoritative DNS server a long with HTTP Server. You may already know what is a
DNS server, if not go to [Wikipedia](https://en.wikipedia.org/wiki/Name_server). What is kind of a new
here is the HTTP server. The HTTP server only provides an API so you can update the DNS zone using a
REST API.

**NOTE**: This is project in a really early stage and it has been build for an specific purpose. That means this software may do not work as expected.
It shouldn't be use in production enviromentes. Use this software at your own risk I am not responsable of anything good or bad thing could happen by using this software.

## DNS Server

The DNS server only understands DNS queries of type **A**.

## HTTP REST API

The API is quite simple, at least at this time.

### CRUD
#### Create

`POST /create/{domain}/{type}/{value}/{ttl}`

```
curl -XPOST http://oscar:8080/create/google.es/A/1.2.3.4/3600
```


#### Read

`GET /read/{domain}/{type}`

```
curl http://oscar:8080/read/google.es/A
```

#### Update

`PUT /update/{domain}/{type}/{value}/{ttl}`

```
curl -XPUT http://oscar:8080/update/google.es/A/4.3.2.1/3600
```

#### Delete

`DELETE /delete/{domain}/{type}`

```
curl -X DELETE http://oscar:8080/delete/google.es/A
```

