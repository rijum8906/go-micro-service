# Core Packages

This package provides core functionality for this project like database connections, token management, some constants that's going to be used by all the services etc

## AppError

- `AppError` is a custom error type that wraps errors from the application.
- It has custom details named field for the error details
- It has feature for auto logging INTERNAL and THIRD_PARTY errors

## Broker

- `broker` is a package that provides structs and Interfaces for message broker management.
- `Client` that holds the NATS connection and provides methods for publishing and subscribing to streams.
- `Consumer`, `Publisher` and `Stream` are the main nterfaces that provide the message broker management functionality.
- This uses NATS with JetStream.

## Cache

- `cache` is a package that provides connection functionality for Redis.

## Constants

- `constants` is a package that provides constants that are used across the application.

## CoreEnv

- `coreenv` is a package that provides the core or minimal environment configuration for a service.
- It has in built support for environment variable parsing and validation.

## CoreLogger

- `corelogger` is a package that provides logger initialization functionality for development, test and production environments.

## CoreOpenFGA

- `coreopenfga` is a package that provides OpenFGA client initialization functionality and OpenFGA model, store and tuple management.

## CoreUtils

- `coreutils` is a package that provides utility functions that are used across the application.

## DB

- `db` is a package that provides database connection functionality (postgres).

## DTO

- `dto` is a package holds some common DTO that's going to use across the application.
- It also has structs uses for sending email data. Has some basic email_template structs.

## Hash

- `hash` is a package that provides hash functions for password hashing and verification.

## Jobs

- `jobs` is a package that provides job names.

## Mailer

- `mailer` is a package that provides email sending functionality using Go Mail.

## Metadata

- `metadata` is a package that provides functions for setting and retrieving metadata from the request context.
- only for grpc requests.

## Permissions

- `permissions` is a package that provides functions for checking permissions using OpenFGA.

## ProtoUtils

- `protoutils` is a package that provides utility functions for working with protobuf messages like parsing, serialization and validation.

## Template

- `template` is a package that provides functions for rendering email templates using Go HTML templates.

## TestUtils

- `testutils` is a package that provides utility functions for testing. (GenerateRandomEmail, MustConnectDB etc.)

## Token

- `token` is a package that provides functions for generating and validating JWT tokens. (auth token and scoped token)
