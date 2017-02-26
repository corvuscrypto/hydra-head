# hydra-head
This is the complement to the hydra data service. The name is lame. I blame Greek mythology for not giving an awesome
name to the immortal head of the hydra after it was severed.

Oh well.

# Configuration

Configuration follows the yaml requirements as in the main hydra program. However for the config of hydra heads
you need to now specify the hydra master address and port so that slave discovery can occur.

That's all for now. OH and don't forget to put the shared OOB secret into the config
