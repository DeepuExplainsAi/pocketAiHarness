package ai.pocketharness.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory

/**
 * Until the app installs and launches the bridge itself (milestone 3), the bridge address,
 * token and workspace are entered by hand from the line `pah-bridge` prints on start.
 */
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent { MaterialTheme { Root() } }
    }
}

@Composable
private fun Root() {
    var listen by remember { mutableStateOf("127.0.0.1:") }
    var token by remember { mutableStateOf("") }
    var cwd by remember { mutableStateOf("") }
    var provider by remember { mutableStateOf("claude") }
    var started by remember { mutableStateOf(false) }

    if (!started) {
        Column(Modifier.fillMaxSize().padding(16.dp)) {
            Text("Pocket AI Harness", style = MaterialTheme.typography.headlineSmall)
            OutlinedTextField(listen, { listen = it }, label = { Text("Bridge address (loopback)") })
            OutlinedTextField(token, { token = it }, label = { Text("Token") })
            OutlinedTextField(cwd, { cwd = it }, label = { Text("Workspace path") })
            OutlinedTextField(provider, { provider = it }, label = { Text("Provider: claude or agy") })
            Button({ started = true }, enabled = token.isNotBlank() && cwd.isNotBlank() && listen.startsWith("127.0.0.1:")) { Text("Connect") }
        }
        return
    }

    val vm: ChatViewModel = viewModel(factory = viewModelFactory {
        initializer { ChatViewModel(BridgeClient(listen, token), provider, cwd) }
    })
    val state by vm.state.collectAsState()
    val online by vm.online.collectAsState()
    ChatScreen(state, online, vm::send, vm::stop, vm::decide)
}
