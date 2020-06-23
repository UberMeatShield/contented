# Contented (Angular and GoBuffalo Setup / Playground)

This is just a simple app for iterating over directories of images given a path.

Mostly it is an example code playground

## Database Setup

It looks like you chose to set up your application using a database! Fantastic!

The first thing you need to do is open up the "database.yml" file and edit it to use the correct usernames, passwords, hosts, etc... that are appropriate for your environment.

You will also need to make sure that **you** start/install the database of your choice. Buffalo **won't** install and start it for you.

### Create Your Databases

Ok, so you've edited the "database.yml" file and started your database, now Buffalo can create the databases in that file for you:

	$ buffalo pop create -a

## Starting the Application

The app is currently split into an angular setup, a gulp helpfile (in Typescript) and the buffalo system.

    $ yarn install
    $ yarn run gulp buildDeploy  # Running just yarn run gulp will kick off the dev builds and watchers
	$ export DIR=/path/containing/image_directories && buffalo dev

## What Next?

Help files, API docs etc

wasd => Navigate up and down
f => Fully load all the directory contents
x => Download. 
e => Fullscreen
q => Close Fullscreen

[Powered by Buffalo](http://gobuffalo.io)
