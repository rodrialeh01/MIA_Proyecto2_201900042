import axios from 'axios';

const instance = axios.create(
    {
        baseURL: 'http://localhost:3000/api',
        timeout: 600000,
        headers: {
            'Content-Type': 'application/json',
        }
    }
); 

export const parse = async (value) => {
    console.log(value);
    const { data } = await instance.post('/consola', { comando: value });
    return data;
}

export const login = async (usuario, id, password) => {
    console.log(usuario, id, password);
    const { data } = await instance.post('/login', { id_particion: id, usuario: usuario, password: password });
    console.log(data);
    return data;
}

export const reportes = async () => {
    const { data } = await instance.get('/reportes');
    return data;
}