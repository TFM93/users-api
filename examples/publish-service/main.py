import os
from google.cloud import pubsub_v1

project_id = os.getenv("PUBSUB_PROJECT_ID", "users-project")
topic_id = os.getenv("PUBLISH_TOPIC_ID", "users")
emulator_host = os.getenv("PUBSUB_EMULATOR_HOST", "localhost:8085")

print(f"HOST {emulator_host} | PROJECT {project_id}", flush=True)
# in case it isn't set
os.environ["PUBSUB_EMULATOR_HOST"] = emulator_host
publisher = pubsub_v1.PublisherClient()

# The full topic name, including the emulator's project ID
topic_path = publisher.topic_path(project_id, topic_id)


def callback(message):
    print(f"Received message: {message.data.decode('utf-8')}", flush=True)
    message.ack()


def publish_message():
    # The message data must be a bytestring
    message_data = "Hello, Pub/Sub Emulator!".encode("utf-8")

    # Publish a message
    future = publisher.publish(topic_path, message_data)
    print(f"Published message ID: {future.result()}")


if __name__ == "__main__":
    publish_message()
