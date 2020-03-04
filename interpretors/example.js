//define an interpretor

module.exports = {
	on_receive : (handler) => {
		//call handler(content, requirement, Id)
	},
	receive : (result, id) => {
	
	},
	run : (content, handler) = > {
		//run content code
		//then call handler(result)
	},
	
	free : () => {
		//true or false
	},
	depedencies_checker : (requirements) => {
		//true or false
		//check if requirements are fulfilled
	}
}
