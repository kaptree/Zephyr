import { ref, onMounted, onUnmounted } from 'vue';

export interface EditingState {
  noteId: string;
  userId: string;
  name: string;
}

export interface NoteUpdatedEvent {
  noteId: string;
  action: string;
  userId: string;
  name: string;
}

export function useGroupSocket(groupId: string) {
  const ws = ref<WebSocket | null>(null);
  const connected = ref(false);
  const editingNotes = ref<Map<string, EditingState>>(new Map());
  const onNoteUpdated = ref<((e: NoteUpdatedEvent) => void) | null>(null);

  function connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const token = localStorage.getItem('auth_token') || '';
    const url = `${protocol}//${host}/ws/group/${groupId}?token=${encodeURIComponent(token)}`;

    const socket = new WebSocket(url);
    ws.value = socket;

    socket.onopen = () => {
      connected.value = true;
      socket.send(JSON.stringify({ event: 'room:join' }));
    };

    socket.onclose = () => {
      connected.value = false;
    };

    socket.onmessage = (ev) => {
      try {
        const data = JSON.parse(ev.data);
        const eventType = data.event;
        if (eventType === 'note:editing') {
          const key = data.note_id;
          const next = new Map(editingNotes.value);
          next.set(key, { noteId: data.note_id, userId: data.user_id, name: data.name });
          editingNotes.value = next;
        } else if (eventType === 'note:idle') {
          const key = data.note_id;
          const next = new Map(editingNotes.value);
          next.delete(key);
          editingNotes.value = next;
        } else if (eventType === 'note:updated') {
          onNoteUpdated.value?.({
            noteId: data.note_id,
            action: data.action || 'updated',
            userId: data.user_id,
            name: data.name,
          });
        }
      } catch {
        /* ignore malformed messages */
      }
    };
  }

  function sendEditing(noteId: string) {
    if (ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ event: 'note:editing', note_id: noteId }));
    }
  }

  function sendIdle(noteId: string) {
    if (ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ event: 'note:idle', note_id: noteId }));
    }
  }

  function sendNoteUpdated(noteId: string, action: string) {
    if (ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ event: 'note:updated', note_id: noteId, action }));
    }
  }

  onMounted(() => {
    connect();
  });

  onUnmounted(() => {
    ws.value?.close();
    ws.value = null;
  });

  return { connected, editingNotes, onNoteUpdated, sendEditing, sendIdle, sendNoteUpdated };
}
