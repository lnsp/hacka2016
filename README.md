# ![honeypot_text_small](https://cloud.githubusercontent.com/assets/3391295/20314615/15e687f6-ab5b-11e6-9385-e0e601800a5b.png)

**Backend service for Honeypot - a HackaTUM 2016 project**

## Endpoints
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Retrieve version information |
| `/register?device=[UUID]&name=[Username]` | GET | Create account with profile name and associated device |
| `/profile?token=[AuthToken]` | GET | Retrieve profile with associated auth token |
| `/profile/{ID}?token=[AuthToken]` | GET | Retrieve profile with associated profile ID |
| `/meet/{UUID}?token=[AuthToken]` | GET | Meet person with associated device UUID |
| `/nearby/{latitude}/{longitude}?token=[AuthToken]` | GET | Retrieve nearby users |
| `/capture/{ssid}?token=[AuthToken]` | GET | Capture hotspot with associated session ID |
| `/settings/name?name=[NewName]&token=[AuthToken]` | GET | Change associated profile name ||
| `/settings/color?color=[NewColor]&token=[AuthToken]` | GET | Change associated profile color |
| `/settings/picture?token=[AuthToken]` | POST | Upload profile picture |
| `/settings/picture/{ID}?token=[AuthToken]` | GET | Retrieve associated profile picture |
| `/hotspot/setup?secret=[SetupSecret]` | GET | Generate new Hotspot credentials |
| `/hotspot/fetch?token=[AuthToken]` | GET | Fetch hotspot state |
| `/hotspot/update?token=[AuthToken]` | GET | Update current session ID |

