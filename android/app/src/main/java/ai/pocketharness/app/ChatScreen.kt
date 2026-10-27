package ai.pocketharness.app

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

@Composable
fun ChatScreen(state: ChatState, online: Boolean, onSend: (String) -> Unit, onStop: () -> Unit, onDecide: (String, Boolean) -> Unit) {
    var draft by remember { mutableStateOf("") }
    Column(Modifier.fillMaxSize().padding(12.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(if (online) "Connected" else "Bridge not connected", style = MaterialTheme.typography.labelMedium)

        LazyColumn(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            items(state.items) { ItemView(it) }
        }

        state.permissions.forEach { PermissionCard(it, onDecide) }

        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            OutlinedTextField(draft, { draft = it }, Modifier.weight(1f), placeholder = { Text("Message") }, enabled = online)
            if (state.status == StreamStatus.Streaming) {
                OutlinedButton(onStop) { Text("Stop") }
            } else {
                Button({ onSend(draft); draft = "" }, enabled = online && draft.isNotBlank()) { Text("Send") }
            }
        }
    }
}

@Composable
private fun ItemView(item: Item) {
    when (item) {
        is Item.User -> Text("You: ${item.text}")
        is Item.Notice -> Text(item.text, color = MaterialTheme.colorScheme.error)
        is Item.Agent -> Column {
            Text(item.text)
            when (item.status) {
                StreamStatus.Interrupted -> Text("Reply interrupted. It may be incomplete.", color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.labelSmall)
                StreamStatus.Failed -> Text("Reply failed.", color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.labelSmall)
                else -> Unit
            }
        }
    }
}

@Composable
private fun PermissionCard(p: PendingPermission, onDecide: (String, Boolean) -> Unit) {
    Card(Modifier.fillMaxWidth()) {
        Column(Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text("The agent wants to use ${p.tool}", style = MaterialTheme.typography.titleSmall)
            if (p.description.isNotBlank()) Text(p.description)
            if (p.input.isNotBlank()) Text(p.input, style = MaterialTheme.typography.bodySmall)
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedButton({ onDecide(p.request, false) }) { Text("Deny") }
                Button({ onDecide(p.request, true) }) { Text("Allow once") }
            }
        }
    }
}
