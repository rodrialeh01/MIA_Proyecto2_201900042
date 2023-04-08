import { createBrowserRouter } from 'react-router-dom';

import Consola from '../pages/Consola';
import Login from '../pages/Login';
import Reportes from '../pages/Reportes';
import NotFound from '../pages/NotFound';


export const router = createBrowserRouter([
    {
        path: '/',
        element: <Consola/>,
        errorElement: <NotFound/>,
    },
    {
        path: '/consola',
        element: <Consola/>,
        errorElement: <NotFound/>,
    },
    {
        path:'/login',
        element: <Login/>,
        errorElement: <NotFound/>,
    },
    {
        path:'/reportes',
        element: <Reportes/>,
        errorElement: <NotFound/>,
    }
]);