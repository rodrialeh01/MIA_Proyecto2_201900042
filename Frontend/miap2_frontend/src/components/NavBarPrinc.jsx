import { NavLink } from "react-router-dom";
import './NavBarPrinc.css';
import { useNavigate } from "react-router-dom";

const NavbarPrinc = () => {
    const navigate = useNavigate();

    const IniciarSesión = () => {
    
        navigate("/login");
    }
    
    return (
        <nav className="navbar bg-body-tertiary" data-bs-theme="dark">
            <div className="container-fluid">
                <a className="navbar-brand fuente"> <img src="src/assets/images/technology.png" alt="" />    DiskMIA - File Command System</a>
                <button className="button_temp" type="submit"onClick={IniciarSesión}><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-person-circle" viewBox="0 0 16 16">
  <path d="M11 6a3 3 0 1 1-6 0 3 3 0 0 1 6 0z"/>
  <path fill-rule="evenodd" d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8zm8-7a7 7 0 0 0-5.468 11.37C3.242 11.226 4.805 10 8 10s4.757 1.225 5.468 2.37A7 7 0 0 0 8 1z"/>
</svg> Inicia Sesión</button>

            </div>
        </nav>
    );
};

export default NavbarPrinc;