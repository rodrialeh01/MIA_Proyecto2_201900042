import React from 'react';
import './login.css';
import { Link } from 'react-router-dom';
import Service from "../Services/Service";
import { useNavigate } from 'react-router-dom';

const Login = () => {

    const [loguear, setLoguear] = useStare({
        id_particion: '',
        usuario: '',
        password: ''
    })
    const id_text = useRef(null);
    const user_text = useRef(null);
    const password_text = useRef(null);

    const handleChange = e => {
        setLoguear({
            ...loguear,
            [e.target.name]: e.target.value
        })
    }

    const handlerLogin = () => {
        Service.login(loguear.usuario,loguear.id_particion,loguear.password)
        .then(({respuesta}) => {
            if (respuesta){
                alert("Bienvenido " + loguear.usuario)
                Navigate('/reportes')
            }else{
                alert("Lo lamento, no puedes iniciar sesión, verifica tus credenciales")
            }
        })
    }

    return (
        <>
            <div className="body_div">
                <div className="session">
                    <div className="left">
                    </div>
                    <form action="" className="log-in" autocomplete="off"> 
                    <h4><img src="src/assets/images/technology.png" alt="" /> <span>Login</span></h4>
                    <div className="floating-label">
                        <input 
                            placeholder="ID Partición" 
                            type="text" 
                            name="id_particion" 
                            id="id_particion" 
                            autocomplete="off" 
                            value={loguear.id_particion}
                            onChange={handleChange}
                        />
                        <label for="id_particion">ID Partición:</label>
                        <div className="icon">
                        </div>
                    </div>
                    <div className="floating-label">
                        <input 
                            placeholder="User" 
                            type="text" 
                            name="usuario" 
                            id="user" 
                            autocomplete="off"
                            value={loguear.usuario}
                            onChange={handleChange}
                        />
                        <label for="user">User:</label>
                        <div className="icon">
                        </div>
                    </div>
                    <div className="floating-label">
                        <input 
                            placeholder="Password" 
                            type="password" 
                            name="password" 
                            id="password" 
                            autocomplete="off"
                            value={loguear.password}
                            onChange={handleChange}
                        />
                        <label for="password">Password:</label>
                        <div className="icon">
                        </div>                        
                    </div>
                    <button type="submit" onClick={handlerLogin}>Log in</button>
                    <Link to="/" className="discrete">Regresar</Link>
                    </form>
                </div>
                </div>
        </>
    ); 
};

export default Login;