const TaskState = {
	Working: 0,
	Done: 1
}
const csrfToken = document.getElementsByName("gorilla.csrf.Token")[0].value
const axiosClient = axios.create({
	  timeout: 1000,
	  headers: { "X-CSRF-Token": csrfToken }
})

const app = new Vue({
	el: '#app',
	data: {
		tasks: [],
		statusOptions: [
			{ value: -1, label: '全て' },
			{ value: 0, label: '未完了' },
			{ value: 1, label: '完了' },
		],
		searchState: -1,
		searchQuery: ''
	},
	methods: {
		checkboxId(id) {
			return "checkbox-" + id
		},
		addTask(event, value) {
			const title = this.$refs.title
			if (!title.value.length) {
				return
			}
			const params = new URLSearchParams()
			params.append('title', title.value)
			params.append('state', TaskState.Working)
			axiosClient.post('/app/tasks', params)
			.then((res) => {
				this.tasks.push({
					id: res.data.created_task.id,
					state: res.data.created_task.state,
					title: res.data.created_task.title
				})
				title.value = ''
			})
			.catch((err) => {
				console.log(err)
				alert("通信エラーが発生しました。もう一度やり直してみてください。")
			})
		},
		updateTask(task) {
			const params = new URLSearchParams()
			params.append('title', task.title)
			if (task.state) {
				params.append('state', TaskState.Done)
			} else {
				params.append('state', TaskState.Working)
			}
			axiosClient.put('/app/tasks/' + task.id, params)
			.catch((err) => {
				console.log(err)
				alert("通信エラーが発生しました。もう一度やり直してみてください。")
			})
		},
		searchTask(state, q) {
			const params = {q: this.searchQuery, state: this.searchState}
			axiosClient.get('/app/tasks', {params: params})
			.then((res) => {
				this.tasks = []
				res.data.tasks.forEach((e) => {
					this.tasks.push({
						id: e.id,
						state: e.state,
						title: e.title
					})
				})
			})
			.catch((err) => {
				console.log(err)
				alert("通信エラーが発生しました。もう一度画面を開き直してみてください")
			})
		}
	},
	computed: {
		labels() {
			return this.statusOptions.reduce(function(acc, e) {
				return Object.assign(acc, { [e.value]: e.label})
			}, {})
		}
	},
	watch: {
		searchState(val, old) {
			this.searchTask(this.searchState, this.searchQuery)
		},
		searchQuery(val, old) {
			this.searchTask(this.searchState, this.searchQuery)
		}
	},
	created() {
		this.searchTask(this.searchState, this.searchQuery)
	}
})
