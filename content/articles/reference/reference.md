+++
hook = "This is my first blog post"
published_at = 2021-04-01T14:52:52Z
title = "Reference post"
tags = ["Test"]
attributions = "Header image by <strong><a href=\"https://www.flickr.com/photos/67499195@N00/717747166\">Andreas Levers</a></strong>. Licensed under Creative Commons BY-NC 2.0."
image = "./pexels-photo-1108099.jpeg"
+++

*Author's note:* This is just my first blog post, testing out all the features in one post. Jump to [What now?](#what-now)

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

This is a [link](https://google.com)


This is a [relative link](/about)

![](./pexels-photo-1108099.jpeg)
*Some puppies*


[1] I realize that REST is designed to provide much greater
    facilities in the form of discovery and content
    negotiation, but in practice these just don't see a lot
    of use, which is why I normally say that convention is
    REST's strongest attribute. [Google.com](https://google.com) is where I go `code it up`
