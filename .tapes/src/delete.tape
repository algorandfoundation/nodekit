Output ./artifacts/nodekit.delete.gif

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


Wait+Screen /Nodekit/

Enter
Show
Sleep 1s
Enter

Type d
Screenshot ./artifacts/nodekit.delete.png

Sleep 3s

Type y
Sleep 1s