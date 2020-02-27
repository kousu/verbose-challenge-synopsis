# Player Developer tech assignment

## Problem description

Imagine you want to update the software of thousands of music players that are already in the field. A music player is composed of multiple components, each having its own version.

We know that every 15 minutes, each player queries an API to see if a new version is available and then updates itself.

## The assignment

You need to create a production-ready tool that will automate the update of a thousand music players by using an API. You don't have to create the API.

Your tool will be used by different people using different operating systems. The most common ones will be Windows, MacOS and Linux.

The input is a .csv file containing, at the very minimum, MAC addresses of players to update, always in the first column.

### Example of a .csv file:
```
mac_addresses, id1, id2, id3
a1:bb:cc:dd:ee:ff, 1, 2, 3
a2:bb:cc:dd:ee:ff, 1, 2, 3
a3:bb:cc:dd:ee:ff, 1, 2, 3
a4:bb:cc:dd:ee:ff, 1, 2, 3
```

The `id1`, `id2` and `id3` fields aren't used in this assignment. The example is shown simply to demonstrate what the .csv file should look like.

### The API to use to update the software version

```
PUT /profiles/clientId:{macaddress}
```

You will need to provide a client id and an authentication token in the headers. For the purpose of this test, these values can be anything but keep in mind that the token will expire in real life.

#### Request

```
Headers

Content-Type: application/json
x-client-id: required
x-authentication-token: required

Body

{
  "profile": {    
    "applications": [
      {
        "applicationId": "music_app"
        "version": "v1.4.10"
      },
      {
        "applicationId": "diagnostic_app",
        "version": "v1.2.6"
      },
      {
        "applicationId": "settings_app",
        "version": "v1.1.5"
      }
    ]
  }
}
```

#### Reponses

##### 200
```
Headers

Content-Type: application/json

Body

{
  "profile": {    
    "applications": [
      {
        "applicationId": "music_app"
        "version": "v1.4.10"
      },
      {
        "applicationId": "diagnostic_app",
        "version": "v1.2.6"
      },
      {
        "applicationId": "settings_app",
        "version": "v1.1.5"
      }
    ]
  }
}
```

##### 401

```
Headers

Content-Type: application/json

Body

{
  "statusCode": 401,
  "error": "Unauthorized",
  "message": "invalid clientId or token supplied"
}
```

##### 404

```
Headers

Content-Type: application/json

Body

{
  "statusCode": 404,
  "error": "Not Found",
  "message": "profile of client 823f3161ae4f4495bf0a90c00a7dfbff does not exist"
}
```

##### 409

```
Headers

Content-Type: application/json

Body

{
  "statusCode": 409,
  "error": "Conflict",
  "message": "child \"profile\" fails because [child \"applications\" fails because [\"applications\" is required]]"
}
```

##### 500

```
Headers

Content-Type: application/json

Body

{
  "statusCode": 500,
  "error": "Internal Server Error",
  "message": "An internal server error occurred"
}
```

## What we mean by production-ready:

- developer documentation (how to build, how to run tests)
- user documentation (how to use the tool, it can be "embedded" in the tool itself or in a document)
- unit tests

## Important notes

- you can use the language/technology of your choice
- explain your assumptions and technical decisions in a document (plain text or markdown, nothing fancy)

## Submission

Submit your assignment using a public GitHub / GitLab / Bitbucket repository (don't use our company name in the repo name or description) or a zip archive.
