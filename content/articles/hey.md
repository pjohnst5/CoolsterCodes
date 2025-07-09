+++
hook = "This is my first blog post"
image = "/content/images/two-phase-render/vista.jpg"
location = "Utah"
published_at = 2025-05-19T14:52:52Z
title = "Ballin' like a baller"
tags = ["Georgia Tech", "AI", "Ballin' it up"]
+++

*Author's note:* This is a longer piece that starts off with exposition into the nature of the N+1 query problem. If you're already well familiar with it, you may want to skip my description of N+1 to a story involving a creative use of [What now](#what-now) to try and plug this hole, or the [two-phase load and render](#two-phase) that I've put in my current company's Go codebase, a pattern we've been using for two years now that's rid of us N+1s, and for which I'd have trouble citing any deficiency (aside from Go's normal trouble with verbosity). It works.

---

## This is a header

And code `two lines after it`
```ruby
ey yo whats up
```

> Sometimes people ask me, why am I so cool?
> 
> Then I answer, just cuz

## What now?

Let's say we have a model `Product` that can render [1] a public-facing API resource for itself by implementing `#render`. I'll be talking about API resources a lot because that's what I'm used, but keep in mind that this could also be an object that's used to render an HTML view and all the same concepts apply.
``` ruby
class Product < ApplicationRecord
  belongs_to :owner # needs to lazy load an owner

  def render
    {
      id:          self.id,
      name:        self.name,
      owner_id:    self.owner_id,
      owner_email: self.owner.email,
    }
  end
end
```

<img src="/content/images/two-phase-render/n_plus_one.svg" alt="N+1.">

This is a [link](https://google.com)


This is a [relative link](/about)

![](/content/images/hey/new_profile_seattle.jpg)
*Me being a baller*


[![](/content/images/hey/new_profile_seattle.jpg)](https://google.com)

[![](/content/images/hey/new_profile_seattle.jpg)](https://google.com)
*this has a caption*

![](/content/images/hey/pexels-photo-1108099.jpeg)
*back to normal*

[1] I realize that REST is designed to provide much greater
    facilities in the form of discovery and content
    negotiation, but in practice these just don't see a lot
    of use, which is why I normally say that convention is
    REST's strongest attribute. [Google.com](google.com) is where I go `code it up`
