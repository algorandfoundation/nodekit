Output ./artifacts/nodekit.lagging.gif
Set Shell "bash"
Set FontSize 19
Set Width 1280
Set Height 640

Require nodekit

Source ./src/theme.tape

Hide
Type "nodekit" Sleep 500ms Enter
Wait+Screen /Nodekit/

Show
Wait+Screen /Out of Sync/
Screenshot ./artifacts/nodekit.lagging.png

Sleep 1.5s Type Y
Wait+Screen /Fast Catchup/
Sleep 3s
Screenshot ./artifacts/nodekit.catchup.png
Ctrl+C
