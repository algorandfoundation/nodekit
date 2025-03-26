Output ./artifacts/nodekit.online.gif

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
Wait+Screen /Nodekit-dev/
Enter Sleep 500ms
Enter Sleep 500ms
Show

Type "r" Sleep 500ms

Wait+Screen /Register Online/
Screenshot ./artifacts/nodekit.online.info.png

Sleep 3s