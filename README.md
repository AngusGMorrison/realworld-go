# _The_ Type-Driven, Hexgonally Architected, Production-Ready, Golang Modular Monolith Example App

![A gopher in priest's robes radiating light](assets/holy-gopher.webp "Let there be light!")

A growing example of a real-world Golang web application, using the best of type-driven design
and Hexagonal Architecture.

This codebase showcases advanced techniques for building production-ready applications that are used daily at leading
tech firms, but are hard to find examples of in the wild.

These patterns will help you slash production bugs, massively increase your test surface, and painlessly evolve your
services as your product grows. Your code will become more readable, more maintainable, and more fun to work on.

## Who is this for?

You'll find this repo valuable if:
* you're excited about Hexagonal Architecture from what you've learned from Uber and Netflix, but aren't sure how to
  apply it to your own projects;
* you're inspired by the promise of type-driven design, which makes invalid data extremely hard to represent;
* you're not ready for the cost and complexity of microservices, but you want an architecture that will seamlessly
  decompose to microservices as you scale;
* you're tired of fixing production bugs, but your current architecture is too hard to test thoroughly;
* you'd like a playground to experiment with microservices patterns, without having to spin up multiple services;

## Why was this built?

It's not easy to find examples of production-grade web applications in the Go ecosystem. It's even harder to find
comprehensive implementations of patterns like Hexagonal Architecture. It's even worse if you're passionate
about using the Go type system to its fullest, catching many common bugs at compile time. Finding an example of all
three in one place? Near impossible.

For years, I've helped industry-leading companies like [Qonto](https://qonto.com/en) develop the architecture
standards used by hundreds of Go engineers. I've introduced or promoted the same techniques at companies as diverse as
ISPs and tier-one investment banks. I've [spoken publicly](https://www.angus-morrison.com/blog/type-driven-design-go)
about the benefits and implementation of type-driven design in Go.

I'd have killed for high-quality examples like this when I started that journey.

In this repo, I've combined everything I've learned about building bomb-proof web applications in Go. I hope you find
these patterns as valuable as I have. If you have questions or feedback, open an issue or drop me a line at
github@angus-morrison.com.

## What does this app do?

This app implements the [RealWorld specification](https://github.com/gothinkster/realworld) for a Medium-like blogging
platform. The RealWorld spec reflects a small but realistic subset of the features you'd find in a genuine product.

This implementation uses Hexagonal Architecture to separate business domains from each other, and from the
infrastructure that ferries data in and out of the application. The example uses JSON over HTTP as its transport layer
and SQLite as its datastore, but our business logic doesn't care. If we moved to gRPC and MongoDB, nothing in the
business logic would need to change. This is an immensely powerful tool for any development team.

But this app goes further by using type-driven design to enforce the validity of data passed into our domains. If a
domain object exists, it's valid. It takes real effort to write code that hands the domain bad data, which means
entire families of sloppy mistakes are eliminated at compile time.

## Hexagawhatnow?

Hexagonal Architecture. You might have heard about it under the name "Ports and Adapters" too.

It was first proposed by [Alistair Cockburn](https://alistair.cockburn.us/hexagonal-architecture/) to address the
symmetrical problems of business logic becoming tangled with the UI layer, and the tight coupling of an application to
its database.

[Netflix](https://netflixtechblog.com/ready-for-changes-with-hexagonal-architecture-b315ec967749) have written about it.
[Uber's Go code structure](https://www.youtube.com/watch?v=nLskCRJOdxM) will also look familiar to Hexagonal
Architecture enthusiasts.

In essence, your business (or "domain") logic – the stuff that makes your product unique – lives at the heart of your
service. It exposes interfaces ("ports") that infrastructural concerns like databases and web servers must conform to
(hence the term "adapters"). Your business logic doesn't care about the nature of the infrastructure it's plugged into,
allowing you to swap out DBs without touching the code that really matters. You can use these same ports to mock
dependencies during testing, granting exceptional coverage using only lightweight unit tests.

Impressively, this pattern has survived the transition from large, enterprise monoliths to modern microservice
architectures. It's that ability to adapt to changing requirements that makes Hexagonal Architecture so powerful.

### But why hexagons?

According to Alistair:

> The hexagon is intended to visually highlight
>
> (a) the inside-outside asymmetry and the similar nature of ports, to get away from the one-dimensional layered picture and all that evokes, and
>
> (b) the presence of a defined number of different ports – two, three, or four (four is most I have encountered to date).

Any nested shape will do. Don't overthink it.

## The _other_ TDD

Type-driven design. The secret sauce that elevates Hexagonal Architecture to Michelin-starred bliss.

Perhaps you've heard the phrase, ["Parse, Don't Validate"](https://lexi-lambda.github.io/blog/2019/11/05/parse-don-t-validate/)?
It expresses the ideal that, rather than rely on runtime validation to ensure the correctness of our data, we can
instead use the type system to guarantee that our data is valid. If it compiles, it's valid.

Wow. Take a moment to let the power of that concept sink in.

Now come back to reality, because the "Parse, Don't Validate" blog post was written about Haskell, and Go's type
system is... well, it's not Haskell's, put it that way.

However, the Go type system has just enough juice that we can make it _hard_ to pass our business logic bad data. Users
can't do it, and developers would have to do it deliberately. Or have a bad accident.

By ensuring that domain endpoints accept only domain models, and by defining constructors for those models such that
bad inputs are always rejected, we can be confident that our business logic is always working with valid data. If
you hold an object of type `Username`, you know it's a valid username. End of.

Not only does this reduce the number of nasty surprises you get in production, it extracts all the noisy validation code
out of your business workflows, and encapsulates it neatly in constructors. This has the triple benefit of making the
validations easy to test, making the business logic easier to read, and reducing the number of ways your business logic
can fail... which makes it easier to test.

If you'd like to learn more about type-driven design in Go, including what happens when you take it to its comical
extreme, check out[ my talk on the subject](https://www.angus-morrison.com/blog/type-driven-design-go) from London
Gophers.

## How this project is structured

### `.`
The project root. Contains `go.mod`, linter and generator config, `Dockerfile` and the root `Makefile`.

### `./assets`
Static assets used by this README. Not part of the application.

### `./bin`
Compiled binaries.

### `./cmd`
The application's entrypoints. Contains the `main` package for each binary, which is responsible for bootstrapping the
application.

#### `./cmd/server`
The entrypoint for the HTTP server. Currently the only entrypoint.

### `./env`
Contains templates for the environment variables required by the build and run phases of the application. The
.env files derived from these templates are .gitignored to protect secrets.

### `./internal`
Library code specific to the application.

#### `./internal/config`
Loads application configuration from the environment.

#### `./internal/domain`
Contains the business logic of the application, with one domain package per business domain. E.g. `./domain/user`, for
all business logic concerning users.

#### `./internal/inbound`
Inbound adapters. These are responsible for:
1. Translating inbound requests from specific transport technologies into
transport-agnostic domain requests.
2. Invoking a domain `Service` instance with the domain request.
3. Translating domain responses back into transport-specific responses.

##### `./internal/inbound/rest`
The REST API adapter that satisfies the RealWorld spec. This spec isn't technically RESTful, but "jsonoverhttp" makes
for a much worse package name.

#### `./internal/outbound`
Outbound adapters. These are responsible for:
1. Implementing outbound ports defined by the domain.
2. Translating requests from the domain into transport or data storage requests.
3. Translating responses from the transport or data storage layer into responses accepted by the domain.

##### `./internal/outbound/sqlite`
This app uses SQLite as its datastore. This package provides the `SQLite` type, which satisfies the `user.Respository`
domain port.

#### `./internal/testutil`
A collection of test utilities useful to all packages.

### `./pkg`
Library code that could be used by other applications, but that I haven't got round to extracting into a separate
repo yet.

### `./scripts`
Build scripts and Docker entrypoints.

### `./tasks`
Makefiles for each major task family, which are imported by the root `Makefile`. This makes managing a large collection
of tasks much easier.

# Progress

Here's what's been implented so far.

## RealWorld Spec

- [x] Users
- [x] Authentication
- [ ] Profiles
- [ ] Articles
- [ ] Comments
- [ ] Tags

## Productionization

- [x] CI pipeline
- [x] Optimized Docker image
- [x] First-class error handling
- [x] Linting
- [x] Extensive, concurrent unit test suite
- [ ] Concurrent integration tests
- [ ] Health checks
- [ ] Structured logging
- [ ] Metrics
- [ ] Tracing
- 