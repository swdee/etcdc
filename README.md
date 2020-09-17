# etcd clientv3 fork

## Overview

This is a fork of github.com/etcd-io/etcd/clientv3 v3.4.13 to make the client its own separate
package compatibile with Go modules, made out of frustration with the upstream project and its handling of modules and GRPC versions seen in the following issues;

https://github.com/etcd-io/etcd/issues/12068  
https://github.com/etcd-io/etcd/issues/12124  
https://github.com/etcd-io/etcd/issues/12276  
https://github.com/etcd-io/etcd/issues/12258  
https://github.com/etcd-io/etcd/issues/11707  
https://github.com/etcd-io/etcd/issues/11721  
https://github.com/etcd-io/etcd/issues/11931  
https://github.com/etcd-io/etcd/issues/11563  

The code base has been further patched to support the `TryLock()` function on `concurrency.Mutex` to allow for non blocking distributed locks from here
https://github.com/etcd-io/etcd/commit/04ddfa8b8dab4ac50dfac8386364d07b79496daf#diff-c43f3df82aca85d77a99d10a4b1c45c6

It requires your application usage to be fixed to GRPC v1.29.1

The purpose of this fork is to get a working etcd client library that works with Go Modules and uses the most recent version of GRPC possible.
This package is meant to be temporary and may become redundant with the release of etcd v3.5 series which aims 
to refactor clientv3 to work with GRPC 1.30 as seen in the
project milestone https://github.com/etcd-io/etcd/milestone/37 in issue https://github.com/etcd-io/etcd/issues/12124

## Install

To use this client in your own application.

Require specified GRPC version

```
go mod edit -replace "google.golang.org/grpc=google.golang.org/grpc@v1.29.1"
```

Install package 

```
go get -u github.com/swdee/etcdc
```


## Usage

Import into your application using a declaration to name as a replacement for upstream clientv3 package.

```
import (
        clientv3 "github.com/swdee/etcdc"
        "github.com/swdee/etcdc/concurrency"
)
```
