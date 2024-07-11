This directory contains sample inputs:

#### `corim.yaml`

This is a YAML representation of a [CoRIM](https://github.com/veraison/corim)
obtained via `cbor2yaml.rb` (see main README).

#### `corim-full.yaml`

Same as `corim.yaml`, except the contained CoMID has also been decoded and
included using `embeddedCBOR` structure (see main README).

#### `ec-p256.jwk`

A key in [JWK](https://www.rfc-editor.org/rfc/rfc7517) format that may be used
to sign a
[COSE_Sign1](https://datatracker.ietf.org/doc/html/rfc8152#section-4.2)
message.

#### `meta.yaml`

An example of additional "meta" data that may be included when generating a
[COSE_Sign1](https://datatracker.ietf.org/doc/html/rfc8152#section-4.2)
message. This is a YAML representation of
[github.com/veraison/corim](https://github.com/veraison/corim) `Meta`
structure.

