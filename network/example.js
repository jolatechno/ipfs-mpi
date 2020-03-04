//define node etc...

module.exports = (args) => return {
	address : () => {
		//return the adress of the node
		return "address";
	},
	send : (adress, message) => {
		//send message to the adress
	},
	pass : (message) => {
		//pass the message to a single random node
	},
	on_receive : (func) => {
		//call func(message) when a message is received
	}
};
