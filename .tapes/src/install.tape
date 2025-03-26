Output ./artifacts/nodekit.install.gif

Set Shell "bash"
Set FontSize 19
Set Width 1280
Set Height 640

Require wget
Require bash
Require sudo

Source ./src/theme.tape

Type@10ms "wget -qO- https://nodekit.run/install.sh | bash" Sleep 100ms Enter

Wait+Screen /Welcome/
Sleep 500ms
Screenshot ./artifacts/nodekit.bootstrap.install.png

Sleep 500ms Type y
Screenshot ./artifacts/nodekit.bootstrap.catchup.png

Sleep 1s Type y

Sleep 3s