Output ../artifacts/bootstrap.gif
Set Shell "bash"
Set FontSize 19
Set Width 1280
Set Height 640

Require nodekit

Source theme.tape



Type "nodekit bootstrap" Sleep 500ms Enter
Wait+Screen /Welcome/
Wait+Screen /Installing/
Screenshot ../artifacts/bootstrap.install.png

Sleep 1.2s Type y
Screenshot ../artifacts/bootstrap.catchup.png
Sleep 1.5s

Ctrl+C
