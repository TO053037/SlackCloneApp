// import TestAPI1 from "./api/test_api_1";
import { User, getUsers, getUserById, postUser, updateUser, deleteUser, } from "./fetchAPI/user";
import testAPI1 from "./api/test_api_1";
export default function Home() {

  console.log(getUsers())
  let user: User = {
    id: "",
    name: "test docker",
    password: "test docker"
  }
  let userRes = postUser(user)
  console.log(userRes)
  return (
    <main>
      <h1>hello nextjs</h1>
    </main>
  )
}