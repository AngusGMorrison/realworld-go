# _The_ Type-Driven, Hexgonally Architected, Production-Ready, Golang Modular Monolith Example App

A growing example of a real-world Golang web application, using the best of type-driven design
and Hexagonal Architecture.

This codebase showcases advanced techniques for building production-ready applications that are used daily at leading
tech firms, but are hard to find examples of in the wild.

These patterns will help you slash production bugs, massively increase your test surface, and painlessly evolve your
services as your product grows. Your code will be become more readable, more maintainable, and more fun to work on.

## Who is this for?

You'll find this repo valuable if:
* you're excited about Hexagonal Architecture by what you've learned from Uber and Netflix, but you're not sure how to
  apply it to your own projects;
* you're inspired by the promise of type-driven design, which makes invalid data extremely hard to represent;
* you're not ready for the cost and complexity of microservices, but you want an architecture that will seamlessly
  decompose to microservices as you scale;
* you're tired of fixing production bugs, but your current architecture is too hard to test thoroughly;
* you'd like a playground to experiment with microservices patterns, without having to spin up multiple services;

## Why was this built?

It's not easy to find examples of production-grade web applications in the Go ecosystem. It's even harder to find
comprehensive implementations of patterns like Hexagonal Architecture. Things are even worse if you're passionate
about using the Go type system to its fullest, catching many common bugs at compile time. You'd have more luck catching
Bigfoot than finding an example of all three in one place.

For several years, I've helped industry-leading companies like [Qonto](https://qonto.com/en) to develop the architecture
standards used by hundreds of Go engineers. I've introduced or promoted the same techniques at companies as diverse as
ISPs and tier-one investment banks. I've [spoken publicly](https://www.angus-morrison.com/blog/type-driven-design-go)
about the benefits and implementation of type-driven design in Go.

I'd have killed for high-quality examples like this when I started that journey.

In this repo, I've combined everything I've learned about building bomb-proof web applications in Go. I hope you find
these patterns as valuable as I have.

## What does this app do?

This app implements the [RealWorld specification](https://github.com/gothinkster/realworld) for a Medium-like blogging
platform. The RealWorld spec reflects a small but realistic subset of the features you'd find in a genuine product.

This implementation uses Hexagonal Architecture to separate business domains from each other, and from the
infrastructure that ferries data in and out of the application. The example uses JSON over HTTP as its transport layer
and SQLite as its datastore, but our business logic doesn't care. If we moved to gRPC and MongoDB, nothing in the
business logic would need to change. This is an immensely powerful tool for any development team.

But this app goes further by using type-driven design to enforce the validity of data passed into our domains. If a
domain object exists, it's valid. It takes real effort to write code that hands the domain bad data, which means a
entire families of sloppy mistakes are eliminated at compile time.

## Hexagawhatnow?

Hexagonal Architecture. You might have heard about it under the name "Ports and Adapters" too.

First proposed by [Alistair Cockburn](https://alistair.cockburn.us/hexagonal-architecture/) as a means to address the
symmetrical problems of business logic becoming tangled with the UI layer, and the tight coupling of an application to
its database, Hexagonal Architecture has been adopted by some of the world's most successful tech companies.

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

# Progress

Here's what's been implemented from the RealWorld spec so far:

- [x] Users
- [x] Authentication
- [ ] Profiles
- [ ] Articles
- [ ] Comments
- [ ] Tags