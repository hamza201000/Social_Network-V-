## Private chat messages

Migration `000008_create_messages` adds the `messages` table with sender and receiver foreign keys to `users(id)`, message text, a seen flag, and a creation time. The existing `followers` table is used to authorize sending: either participant must follow the other. Indexes support loading a conversation in both sender/receiver directions.