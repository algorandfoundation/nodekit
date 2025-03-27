Output ./artifacts/nodekit.create.gif

Set Shell "bash"
Set FontSize 14
Set Width 1280
Set Height 640

Require wget
Require bash
Require sudo

Source ./src/theme.tape

Hide
Type "nodekit" Sleep 100ms Enter

Show
Wait+Screen /Nodekit/

Sleep 500ms Type g

Screenshot ./artifacts/nodekit.create.address.png
Type@10ms "TAPEQCZOD42X3F4KRPVXEYI76V2WYBSWCD7QRZHY7PCLVLX74FKIIXWQHM" Sleep 100ms Enter

Screenshot ./artifacts/nodekit.create.duration.png
Sleep 500ms Type "2"
Sleep 1s Enter

Wait+Screen /Generating Keys/
Screenshot ./artifacts/nodekit.create.waiting.png

Wait+Screen /Key Information/
Screenshot ./artifacts/nodekit.create.info.png
Sleep 3s