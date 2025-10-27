+++
title = "Congratulations"
description = "Contact CoolsterCodes"
hidden = true
+++

You have reached my secret contact page 🏆

In order to reach out to me, run this script and email me at the given address

```bash
cat << EOF | python3
x = b'\xaa\xe3\x35\x16\xe2\xd0\x4b\x8f\xb9\x46\xcf\x9d\x0d\x2e\xcb\xab\x0f\x88\x3f\xfa\xae\x9b\x7d'
y = b'\xc9\x8c\x5a\x7a\x91\xa4\x2e\xfd\xda\x29\xab\xf8\x7e\x6e\xac\xc6\x6e\xe1\x53\xd4\xcd\xf4\x10'
print(''.join(chr(a^b) for a,b in zip(x,y)))
EOF
```
