# upptime-email-notify

Notify via emails using GitHub webhooks (since the built-in ones aren't working). 

# How this works

This is a simple Go server that recieves some data and sends a notification to configured recipients. 

Since if some endpoint goes down Upptime opens an issue on the repository, we can use that issue creation event to send a notification through an external server using webhooks. This kinda defeats the purpose of "server-less uptime monitor", but the built in notification systems aren't working, and we need some way to get notifications and I'm not knowledgable at TS/JS or any web technologies really, this is my way of handling this instead of patching Upptime itself (till the moment Anand (or some other contributor) fixes it).

# How to deploy

You can deploy it in two ways, through direct http like a naked server or behind a reverse proxy. I recommend the second method which uses a reverse proxy and runs under a Docker container.

**For the reverse proxy setup see [my article](https://linuxhandbook.com/nginx-reverse-proxy-docker/) on Linux Handbook.**

## Prerequisites

Before moving forward make sure you have `git`, `docker` and `docker-compose` installed.

## Using docker

**1. Clone the repository**

First clone this repository and `cd` into it

```
git clone https://github.com/debdutdeb/upptime-email-notify.git \
	&& cd upptime-email-notify
```

**2. Set up the environment variables**

Open the `.env` file and assign values to the variables according to the following list

> I have made the Upptime related environment variables similar to Upptime's ones.

- `NOTIFICATION_EMAIL_SENDGRID`: If you're going to use SendGrid, set this to any non empty value.
- `NOTIFICATION_EMAIL_SENDGRID_API_KEY`: Your SendGrid API key (is using Sendgrid).
- `NOTIFICATION_EMAIL_FROM`: The `From` value on your email (like noreply@domain.com).
- `NOTIFICATION_EMAIL_TO`: This is a space separated list of email ids that the emails are going to be sent to.
- `NOTIFICATION_EMAIL_SMTP_HOST` & `NOTIFICATION_EMAIL_SMTP_PORT`: If not using SendGrid, the hostname and port of your SMTP server respectively.
- `NOTIFICATION_EMAIL_SMTP_USERNAME` & `NOTIFICATION_EMAIL_SMTP_PASSWORD`: Username and password for authentication with the SMTP server.
- `GITHUB_SECRET`: This is a random string, used to verify the sender. Use something like `openssl rand -hex 12`. Remember the string since you're going to need this one later.
- `VIRTUAL_HOST` & `LETSENCRYPT_HOST`: See [this](https://linuxhandbook.com/nginx-reverse-proxy-docker/).
- [OPTIONAL] `ENDPOINT`: The URL endpoint, defaults to `/issue`.

> I have not added something like `NOTIFICATION_EMAIL_MAILGUN` since you can use the `NOTIFICATION_EMAIL_SMTP_HOST` family of environ variables to configure that just like you can for SendGrid.

**3. Build the image**

Build the image using the following command

```
docker-compose build
```

**4. Change the network**

Make sure you change the network in `docker-compose.yaml` according to your reverse proxy setup.

**5. Deploy the container**

Deploy the container with 

```
docker-compose up -d
```

## Naked server

If not inclined to use `docker` you can simply do `go get github.com/debdutdeb/upptime-email-notify`, and then run `$(go env GOPATH)/bin/upptime-email-notify`.

# The GitHub repository webhook

1. Go to your repository settings, go to webhooks from the left side panel, click on "Add webhook".

2. Here add your endpoint, if behind a reverse proxy, it's going to be https://domain.com/$ENDPOINT (ENDPOINT being whatever value you set the environment variable to), otherwise it'd be http://ip:8080/$ENDPOINT.

3. For the content type, make sure it is set to `application/json`.

4. Remember the value of `GITHUB_SECRET` environment variable? Paste that in the "Secret" text field.

5. Manually select the event type to be only "issues", nothing else.

6. Check the "Active" checkbox and click on "Add webhook".
