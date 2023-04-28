import React from 'react';
import './login.css';
import { Link } from 'react-router-dom';
import Service from "../Services/Service";
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
// ES6 Modules or TypeScript
import Swal from 'sweetalert2'



const Login = () => {
    const navigate = useNavigate();
    const [loguear, setLoguear] = useState({
        id_particion: '',
        usuario: '',
        password: ''
    })

    const handleSubmit = (e) => {
        e.preventDefault();
    }

    const handleChange = e => {
        setLoguear({
            ...loguear,
            [e.target.name]: e.target.value
        })
    }

    const handlerLogin = () => {
        Service.login(loguear.usuario,loguear.id_particion,loguear.password)
        .then(({autenticado})=>{
            console.log(autenticado)
            if (autenticado){
                let timerInterval
                Swal.fire({
                title: "Bienvenido " + loguear.usuario + "!",
                icon: 'success',
                html: 'Esta bienvenida terminara en <b></b> millisegundos.',
                timer: 1500,
                timerProgressBar: true,
                didOpen: () => {
                    Swal.showLoading()
                    const b = Swal.getHtmlContainer().querySelector('b')
                    timerInterval = setInterval(() => {
                    b.textContent = Swal.getTimerLeft()
                    }, 100)
                },
                willClose: () => {
                    clearInterval(timerInterval)
                }
                }).then((result) => {
                /* Read more about handling dismissals below */
                if (result.dismiss === Swal.DismissReason.timer) {
                    console.log('I was closed by the timer')
                }
                })
                navigate('/reportes')
            }else{
                Swal.fire({
                    title: 'Credenciales Incorrectas!',
                    text: 'Intenta de nuevo',
                    icon: 'error',
                    confirmButtonText: 'Ok'
                })
                setLoguear({
                    id_particion: '',
                    usuario: '',
                    password: ''
                })
            }
        })

    }

    return (
        <>
            <div className="body_div">
                <div className="session">
                    <div className="left">
                    </div>
                    <form action="" className="log-in" autocomplete="off" onSubmit={handleSubmit}> 
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