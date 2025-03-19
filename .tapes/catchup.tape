Output ../artifacts/catchup.gif
Set Shell "bash"
Set FontSize 19
Set Width 1280
Set Height 640

Require nodekit

Source theme.tape

Hide
Type "nodekit" Sleep 500ms Enter


Wait+Screen /Fast Catchup/
Screenshot ../artifacts/catchup.png
Show
Sleep 3s
Ctrl+C
