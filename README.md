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
    // SkipHTTPs: false
    SkipSubPath: true
  - URL: 'www.hotmail.com'
    Disable: true
```
# Run Time
```
www.google.com
test -> http://www.google.com/
test -> http://www.google.com/AA
test -> http://www.google.com/BB

www.yahoo.com
test -> https://www.yahoo.com

www.hotmail.com
test -> not run
