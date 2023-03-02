import React, { useState } from "react";
import { useCookies } from "react-cookie";
import { currentUser, login } from 'pages/fetchAPI/login'
import { getWorkspaces, Workspace } from 'pages/fetchAPI/workspace'

function LoginForm() {
  
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [cookies, setCookie, removeCookie] = useCookies(['token']);
  const [workspacelist, setWorkspaceList] = useState([
    {
      id: 0,
      name: "",
      primary_owner_id: 0
    }
  ]);

  const list = workspacelist.map((item, index) => (
    <div key={index}>
      <p>{item.id}</p>
      <p>{item.name}</p>
      <p>{item.primary_owner_id}</p>
    </div>
  ));

  const nameChange = (e: any) => {
    setName(e.target.value);
  };
  const passwordChange = (e: any) => {
    setPassword(e.target.value);
  };

  const handleLogout = () => {
    console.log("logout");
    removeCookie("token", {path: '/'});
  };
  const handleLogin = () => {
    console.log("login");
    let user = { name: name, password: password }
    login(user).then((currentuser: currentUser) => { 
      setCookie("token", currentuser.token);
    });
    getWorkspaces().then((workspaces: Workspace[]) => { 
      console.log("workspaces")
      console.log(workspaces)
      setWorkspaceList(workspaces)
      console.log(workspacelist)
    });
  };

  return (
    <div className="App">
        <label>名前
          <input type="text" value={ name } name="name" onChange={(e) => nameChange(e)} />
        </label><br />
        <label>パスワード
          <input type="password" value={ password } name="password" onChange={(e) => passwordChange(e)} />
        </label><br />
        <button onClick={handleLogin} >ログイン</button>
      <button onClick={handleLogout}>ログアウト</button>
      <div>
        {list}
      </div>
    </div>
  );
}

export default LoginForm;
