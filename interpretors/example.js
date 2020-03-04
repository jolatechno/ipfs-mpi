//define an interpretor

module.exports = {
	on_receive : (func) => {
		//call func(content, requirement, Id)
	},
	receive : (result, id) => {
	
	},
	run : (content) = > {
		//run content code
	},
	
	free : () => {
		//true or false
	},
	depedencies_checker : (requirements) => {
		//true or false
		//check if requirements are fulfilled
	}
}
