Output ./artifacts/nodekit.offline.gif

Set Shell "bash"
Set FontSize 15
Set Width 1280
Set Height 1000

Require nodekit

Source ./src/theme.tape

Hide
Type "nodekit --force-qr" Enter
Wait+Screen /Nodekit/

Enter Sleep 500ms
Wait+Screen /YES/
Show
Enter Sleep 1s
Type "o" Sleep 1s

Wait+Screen /how QR/
Screenshot ./artifacts/nodekit.offline.lora.png

Type "s" Sleep 1s
Screenshot ./artifacts/nodekit.offline.qr-code.png

Wait+Screen /Successfully/
Screenshot ./artifacts/nodekit.offline.qr-code.success.png
Sleep 2s