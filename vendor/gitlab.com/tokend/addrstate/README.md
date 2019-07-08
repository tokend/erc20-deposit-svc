Addrstate is used to get LedgerEntryChange for specific effects and entryTypes. After creating Watcher gets effect and entryTypes from all mutators which were given to him and start streams transactions fetched for specified filters. It gets updates for all mutators and store them in state. When the data is required it is obtained from getters. Getter checks if latest ledger was reached, gets raw data from state and returns it in convinient form. 

# Add new mutator

To add new mutator create struct in StateUpdate with fields you want to get. Create mutator and implement StateMutator interface with your effects and entryTypes. In State struct add map to store raw data. In Mutate() check if StateUpdate contains your updates and save it to your map. Create getter to handle raw data and return it in convinient form.
