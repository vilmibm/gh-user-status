# gh user-status

being an extension for interacting with the status on a GitHub profile.

- `gh user-status set`
	- `gh user-status set` interactively set status
	- `gh user-status set --limited "vacation"` set a status with limited availability
	- `gh user-status set --expiry 1h "leave me alone"` set with 1 hour expiry
	- `gh user-status set --emoji "pizza" "eating lunch"` set with an emoji
- `gh user-status get`
	- `gh user-status get` see your status
	- `gh user-status get mislav` see another user's status

By default, the :thought_balloon: emoji is used.

Limiting visibility of the status to an organization is not yet supported.

# author

vilmibm <vilmibm@github.com>
