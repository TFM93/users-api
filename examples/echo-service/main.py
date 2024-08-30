import os
from google.cloud import pubsub_v1

project_id = os.getenv("PUBSUB_PROJECT_ID", "users-project")
subscription_id = os.getenv("ECHO_SUBSCRIPTION", "my-subscription")
emulator_host = os.getenv("PUBSUB_EMULATOR_HOST", "localhost:8085")

print(f"HOST {emulator_host} | PROJECT {project_id}", flush=True)
# in case it isn't set
os.environ["PUBSUB_EMULATOR_HOST"] = emulator_host
subscriber = pubsub_v1.SubscriberClient()
subscription_path = subscriber.subscription_path(project_id, subscription_id)


def callback(message):
    print(f"Received message: {message.data.decode('utf-8')}", flush=True)
    message.ack()


subscriber.subscribe(subscription_path, callback=callback)

print(f"Listening for messages on {subscription_path}...", flush=True)

# Keep the service running
while True:
    pass
