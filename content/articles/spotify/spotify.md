+++
title = "How to Program Spotify!"
hook = "Make playlists automatically"
image = "spotify_logo.png"
published_at = 2026-04-13T19:11:54-06:00
tags = ["Music", "Spotify", "Programming"]
youtube = ""
+++

I have a very particular used case where on my Garmin watch that can have up to **2,000** songs downloaded onto it from Spotify.

Here's the catch: **Syncing** that many songs regularly is MASSIVE, and takes forever.

Plus, I usually don't even get CLOSE to listening to that many songs while running. Maybe like 10-20 AT MOST.

So, here's the catch: I want to program a playlist or two, that update periodically to my watch, that are just samples of other playlists.

![](./fenix.jpg)
*My beautiful Garmin Fenix 7S Pro Sapphire Solar watch!! 😱*

## Playlists

- "Random 20" (A playlist of only 20 random songs from my liked songs, that updates once a day)
- "Last 20" (My most recently liked 20 songs, also updated every day)

## Spotify API

Spotify has an API that let's you **code** your playlists [1] 😂 (My brother would be teasing me so bad right now)

![](./Spotify_API.png)
*The beloved [Spotify API](https://developer.spotify.com/documentation/web-api)*

## GitHub

Here is the GitHub repository for this project.

Feel free to clone it and go!!

![](./github_spotify.png)
*The beautiful [source code](https://github.com/pjohnst5/Spotify)*


## Walkthrough of the code

So here's the deal, you could make raw API request to Spotify if you wanted to using any computing language you wanted, but that is quite verbose and unwieldy. There's a nice python package called spot which lets you do this in python very easily.

![](./raw_vs_spotipy.jpg)
*Raw `curl` requests to Spotify's API vs Spotipy*

## Getting a client-id

First you have to get a client-id by registering an app with Spotify. I think you only get one per account unless you pay more or something, but essentially it's letting the app act on your behalf and alter your Spotify library for you if you allow it (the first time you run the app, it will prompt you to allow the app to alter your Spotify account)

![](./appifier.png)
*My first Spotify app*

## Deployment

This code runs once a day, on an [Azure Function](https://azure.microsoft.com/en-us/products/functions) in cron-job type fashion

![](./function_App.png)
*The Azure Function App running my python code every day*

## Monitoring

This Azure Function App prints its logs to an Azure Applciation Insights

![](./Logs.png)
*Example run of this app*

## Cost breakdown

This only cost **$0.05** per MONTH! 😱

![](./cost.png)

[1] Now Spotify makes you have a "Premium" account to do this, but before it used to be possible with a free account

![](./premium.png)
*Spotify being Spotify*

