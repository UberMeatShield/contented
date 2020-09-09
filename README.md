# Contented (Angular and GoBuffalo Setup / Playground)

This is just a simple app for iterating over directories of images given a path.

Mostly it is an example code playground so you can browse images, see Angular keypress navigation events and hack around
with the usual http downloads, headers and ajax support.

## Database Setup

This is done via docker right now, if you just do docker-compose up it will get you postgres, but that is not required.

It looks like you chose to set up your application using a database! Fantastic!

The first thing you need to do is open up the "database.yml" file and edit it to use the correct usernames, passwords, hosts, etc... that are appropriate for your environment.

You will also need to make sure that **you** start/install the database of your choice. Buffalo **won't** install and start it for you.

### Create Your Databases (Not required)

Ok, so you've edited the "database.yml" file and started your database, now Buffalo can create the databases in that file for you:

	$ buffalo pop create -a

    TODO: This is currently disabled as there is no backing DB model yet but that will come later

###  Development

    See a list of all the available tasks
    $ yarn run gulp --tasks 

    For doing rapid development on javascript I run these in two different consoles
    $ yarn run ng build contented --configuration=dev --watch=true --deploy-url /public/build/
    $ yr ng test contented --watch=true

TODO: Make it so the base gulp run is a little smarter, buffalo dev is SUPPOSED to do solid reloads on file
change but it is sketchy.  Using a raw go httpd server worked better using change detection 
TODO: maybe stop using buffalow dev and instead rebuild and kick it manually...

## Starting the Application

The app is currently split into an angular setup, a gulp helpfile (in Typescript) and the buffalo system.

    $ yarn install
    $ yarn run gulp buildDeploy  # Running just yarn run gulp will kick off the dev builds and watchers
	$ export DIR=`pwd`/mocks/content/ && buffalo dev
	$ export DIR=`pwd`/mocks/content/ && buffalo test

## What Next?

Need to add in something around tests and exporting the dir path.

## What this does

This is pretty much just a holder for getting GoLang, GoBuffalo and Angular trying to play together.

Browsee over a set of images (expand to video?)

wasd => Navigate up and down
f => Fully load all the directory contents
x => Download. 
e => Fullscreen
q => Close Fullscreen

[Powered by Buffalo](http://gobuffalo.io)
