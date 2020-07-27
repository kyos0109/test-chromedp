# test-chromedp
Use chromedp test website.

# Config File Example

config.yaml
```
HTML:
  BodyWaitDomLoad: '{JSPath}'
TestConfig:
  SubPath:
    - '/'
    - '/AA'
    - '/BB'
TestTargets:
  - URL: 'www.google.com'
    SkipHTTPs: true
  - URL: 'www.yahoo.com'
    SkipHTTPs: true
    SkipSubPath: true
  - URL: 'www.hotmail.com'
    Disable: true
```
