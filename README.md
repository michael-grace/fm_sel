> [!CAUTION]
> The README is out of date, and needs updating.

## yoooo, it's the FM source selector

yeah, so this is what gets streamed to the FM transmitter, and doesn't consider the fallbacks on the transmitter side

```
                                 jukebox
                                      |                        reserve PGM feed
                                      |
                                      |                            |
                                      |                            |
                                      |                            |
                                      |                            |
                                      |                            |
                                      |                            |
                              +-------v------+                     v
                              |              |               +-----------+
 PGM feed                     |              |               |           |
from studio selector---------->              |   RTP link    |           |
                              |    dolby     +-------------->| fmstl     +---->TX
                              |              |               |           |
                              |              |               |           |
                              +----^---------+               +-----------+
                                   |                              ^
                                   |                              |
                                   |                              |
                                   |                              |
                                   |                              |
                                   |                              |
                                   |                              |
                              autonews
                                                            technical difficulties
                                                           audio
```

### controlling

It supports three sources (tho, could be expanded without too much work)

| ID | Source |
|----|--------|
|0 | PGM Feed |
| 1 | Jukebox |
| 2 | AutoNews |

You can poll the currrent source with `curl localhost:5001/source` (this works from any host).
You can set a source with `curl -X POST -d source=N localhost:5001/source` (this only works from localhost, where `N` is the source number)

The current source is stored in `/opt/fm_sel/source.txt`. It logs to `/opt/fm_sel/sel.log`.