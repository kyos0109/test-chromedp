# test-chromedp
Use chromedp test website.

# Config File Example

config.yaml
```
HTML:
  BodyWaitDomLoad: '{JSPath}'
TestConfig:
  Remote: true
  ChromedpWS: 'ws://127.0.0.1:9222/devtools/page/XXXXXXXXXXXXXXXXXXXXXXXXXXXXX'
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
