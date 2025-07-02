![](/content/images/hey/new_profile_seattle.jpg)

# Sup
I'm Paul. I work full-time for [Microsoft](https://www.linkedin.com/in/paul-alonso-johnston/), and like to make [YouTube videos](https://youtube.com/@coolstercodes) by night.

# Georgia Tech
At this point, 

## Technology (#technology)

This site is a static set of HTML, JS, CSS, and image files built using a [custom Go executable](https://github.com/brandur/sorg), stored on S3, and served by a number of worldwide edge locations by CloudFront to help ensure great performance around the globe. It's deployed automatically by CI as code lands in its master branch on GitHub. The architecture is based on the idea of [the Intrinsic Static Site](/aws-intrinsic-static).

It was previously running [Ruby/Sinatra stack](https://github.com/brandur/org), hosted on Heroku, and using CloudFlare as a CDN.

## Design (#design)

This site was initially designed with a boatload of custom CSS which over the years became a rat's nest in which it was difficult to change anything without breaking something else. I've since simplified things and moved it all over to [Tailwind](https://tailwindcss.com/), now being of the opinion that CSS as a concept is fundamentally unmaintainable.
