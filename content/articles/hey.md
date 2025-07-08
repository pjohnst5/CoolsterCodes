+++
hook = "This is my first blog post"
image = "/content/images/two-phase-render/vista.jpg"
location = "Utah"
published_at = 2025-05-19T14:52:52Z
title = "Ballin' like a baller"
tags = ["Georgia Tech", "AI", "Ballin' it up"]
+++

## This is a header
And code `two lines after it`
```ruby
ey yo whats up
```

## What now?
Let's say we have a model `Product` that can render a public-facing API resource for itself by implementing `#render`. I'll be talking about API resources a lot because that's what I'm used, but keep in mind that this could also be an object that's used to render an HTML view and all the same concepts apply.
```ruby
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
