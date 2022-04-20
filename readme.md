# TODO
- server deployment (vm)
- dead player state (player handlers)
- better serialization (right now it's json - it can be too much data to handle many players)
- score (create score on the server side, bump score when player - ball collision happens - broadcast reliable message when score changes (or use already existing HitConfirmPacket))
- server commands - e.g. end warmup, restart game etc.
- waiting for a game (players can join and get assigned to team and throw balls but after being hit they should be killed and respawned n seconds after) - warmup
- charge / dash cooldown visual indicators
- measurements (networks stats + some other stats + is server really running at rate we've set it to ? or for some reason it runs slower ?)
- graphics improvements e.g.
  - when player hit could set player into dead state and spawn particles that would shoot in direction where ball was going
  - when player hit we could do some screen shake or smth similar (examples at ebiten site)

# Fix
- lagging ball players too - server not being able to hit set 144 tick rate on my pc :/// - probably due to sync map loops
- you can grab the ball from the others (again :[)
- when one player closes the client, server crashes (probably because tcp connection wasn't closed properly)

# Nice to have
- client side prediction
- web client (webRTC?)