---
sidebar_position: 2
title: "002 - API Contracts (OpenAPI)"
---

# [002] Define API contracts through OpenAPI

* Status: accepted
* Author: @antonve
* Date: 2022-12-20

## Context

At the moment we have no standard for defining API contracts. API endpoints are implemented on the backend on the spot, and then immediately implemented on the frontend. This makes things more fragile when changes are needed and lacks documentation about the existing APIs. In other words, backend and frontend are tightly coupled with invisible and undocumented API contracts. We are also unable to generate boilerplate code to implement the server and client, resulting in lots of toil in writing it ourselves.

## Considered options

### gRPC

> gRPC is a modern open source high performance Remote Procedure Call (RPC) framework that can run in any environment. It can efficiently connect services in and across data centers with pluggable support for load balancing, tracing, health checking and authentication. It is also applicable in last mile of distributed computing to connect devices, mobile applications and browsers to backend services.

gRPC is a popular technology to build APIs with a strongly typed contract. It uses protobuf, which is a binary protocol. It supports many languages, plugins, and also bi-directional streaming.

#### Pros

* Lightweight wire-format
* High performance (although in certain in cases protobuf can become a bottleneck)
* Supports code generation for our stack (TypeScript and Go)
* Great integration with Bazel
* Bi-directional streaming in case we want to make the leaderboard experience real-time
* Uses protobuf as the IDL  (Interface Description Language), which makes it difficult to make APIs that break backwards compatibility
* Can be configured in such a way that it also exposes a HTTP Rest API

#### Cons

* A binary wire-format also has the downside that it's not as straightforward to interact with it. You need a gRPC client in order to make requests.
* Calling gRPC APIs from the web is not straightforward. Usually a proxy to translate gRPC to JSON/Rest is required to make it accessible from the browser.
* In general more complex to work with and comes with a steep learning curve

### OpenAPI

A specification that describes how APIs work. It's a yaml file that describes all endpoints. Tools are available to generate boilerplate server/client code in several languages and API documentation.

#### Pros

* Easy to get started with
* No vendor lock in to a specific technology as it uses the most popular and straightforward protocols like JSON and HTTP. 
* A client has the option to use a generated API client or write a one-off query to the API without needing any new dependencies.
* Uses technology most developers are familiar with, so the learning curve is rather low. There's enough tooling available to get started using it without ever having used it.

#### Cons

* Integration for Bazel is a bit annoying as you'll need to manually manage some of the build tools for Golang
* Typing guarantees aren't as strong when compared to gRPC. It's possible to send a response that doesn't meet the API spec.

### GraphQL

I've briefly considered GraphQL to solve this solution, but ultimately found that my experience with it wasn't satisfactory to continue to explore it deeper. In my opinion the biggest problem GraphQL solves is communication across large engineering teams working on related but separate projects. Tadoku will most likely never face this problem and it wouldn't be worth the complexity it introduces on both the client and the server.

## Decision

In the end I decided to go with OpenAPI due the simple nature of it. I've had good experience with gRPC in the past but that came with a steep learning curve and more complex development flow. Some of the amazing gRPC features aren't relevant to Tadoku (low wire format, super fast, streaming) which makes the tradeoff in complexity not worth it.

## Outcome

To be written after the v2 rewrite.
