import React from 'react';
import './login.css';
import { Link } from 'react-router-dom';

const Login = () => {
    return (
        <>
            <div className="body_div">
                <div className="session">
                    <div className="left">
                    </div>
                    <form action="" className="log-in" autocomplete="off"> 
                    <h4><img src="src/assets/images/technology.png" alt="" /> <span>Login</span></h4>
                    <div className="floating-label">
                        <input placeholder="ID Partición" type="text" name="id_particion" id="id_particion" autocomplete="off"/>
                        <label for="id_particion">ID Partición:</label>
                        <div className="icon">
                        </div>
                    </div>
                    <div className="floating-label">
                        <input placeholder="User" type="text" name="user" id="user" autocomplete="off"/>
                        <label for="user">User:</label>
                        <div className="icon">
                        </div>
                    </div>
                    <div className="floating-label">
                        <input placeholder="Password" type="password" name="password" id="password" autocomplete="off"/>
                        <label for="password">Password:</label>
                        <div className="icon">
                        </div>                        
                    </div>
                    <button type="submit" onClick="return false;">Log in</button>
                    <Link to="/" className="discrete">Regresar</Link>
                    </form>
                </div>
                </div>
        </>
    ); 
};

export default Login;