# dolabra

dolabra is a Trip Management System, aimed to be help outdoor
clubs / organizations organize and plan trips. Unlike a generic meetup service
(meetup.com for example), dolabra has many features that would be unused for
non-outdoor trips.


## Configuration

### Photos

Photos are stored on gdrive and require 2 accounts minimum
* One to own the photos and share them with other leaders (through gdrive)
* One system account created with GCP to manage the photos
  * You will be given the option to download system account credentials during
    creation, which are used below

The photos structure is fairly flexible. We only neeed to know the folder
containing trips and the folder containing homephotos

We recommend creating a "main" folder for all photos, then creating 2 folders
inside named `TRIPS` and `HOMEPHOTOS`. Folders are auto-created inside these
folders when photos are uploaded.

### SMTP

The `SMTP_FROM_*` variables are used to create a system account as the first
entry in the database so it's easy to reference when sending emails without
additional overhead.

tl;dr we can pass in `0` as the person who's sending the email and it just
works(tm).

This system account is *always* put in the From field and should be designed to
catch bounced emails. Sometimes the reply-to field is also filled if the email
is being sent by a specific person.

### Dev Mode

The `DEV` variable enables the `/auth/dev/{sub}` path allowing scripts to
easily send in a custom `sub` (auth identifier) without going through a
proper idp.

### Environmental variables

Create `dolabra.env` (copy `dolabra.env.sample`) with the following variables
defined:
* `COOKIE_DOMAIN`: Domain to use for cookies (should be shared between api & frontend)
* `FRONTEND_URL`: Frontend url (for linking from emails)
* `AWS_ACCESS_KEY_ID`: AWS access key (for SES)
* `AWS_SECRET_ACCESS_KEY`: AWS secret key (for SES)
* `GOOGLE_CLIENT_SECRET`: Client Secret for Google Sign-in
* `GOOGLE_CLIENT_ID`: Client Id for Google Sign-in
* `GOOGLE_APPLICATION_CREDENTIALS`: Path to json file with GCP system account
  credentials. Place this file in `utils/` since this folder is shared with
  docker (for Photos)
* `GDRIVE_TRIPS_FOLDER_ID`: Gdrive folder containing trips (for Photos)
* `GDRIVE_HOME_PHOTOS_FOLDER_ID`: Gdrive folder containing homephotos
  (for Photos)
* `SMTP_FROM_FIRST_NAME_DEFAULT`: Firstname of person in the `From` field
* `SMTP_FROM_LAST_NAME_DEFAULT`: Lastname of person in the `From` field
* `SMTP_FROM_EMAIL_DEFAULT`: Email put in the `From` field
* `STRIPE_PUBLIC_KEY`: Public key for Stripe payments
* `STRIPE_SECRET_KEY`: Secret key for Stripe payments
* `STRIPE_WEBHOOK_SECRET`: Webhook secret for Stripe payments
* `DEV`: Optionally set to `1` to enable developer moded


## Running it

`launch.sh` starts a local instance for dev or testing purposes. Take a look at
[our docker repo](https://gitlab.com/ocvt/docker) for examples running in
production.


## Testing

* `make static-check` runs `go vet` and `sqlvet`
* `make format` runs `go fmt`
* `make integration-test` builds everything and runs python integration tests
  from the `tests` folder


### Notes

* 403 is returned if the error is related to permissions or user input (ie trying to signup with a
  pet on a trip that does not allow pets), otherwise 400 is used for generic errors or whenever we
  don't 100% know the cause of the issue

## TODO

* Put all env vars in config
* Ensure email errors from bad email addresses are dequeued instead of infinitely trying
* Add check in tasks to ensure officers or approvers are removed after expiration
* Ensure errors after row db queries are checked
* Test /trips/myattendance
* Validate trip input data
* organize functions A-Z (approval & notify related files: TODO)
* Test /trips/{tripId}/mystatus
* Ensure all input is sanitized (signup notes, other free text fields)
* Test GetTripMyStatus
* Make tasks work (db locking issue and no error check after rows.Next())
* Use db transactions
* Use proper 1/0 instead of true/false for sqlite
* Add NotFound & MethodNotAllowed handlers
