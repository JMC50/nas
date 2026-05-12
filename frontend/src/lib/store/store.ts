import {writable} from 'svelte/store'

export interface userclass {
    userId: string;
    username: string;
    krname: string;
    global_name: string;
    token: string;
}

const userDetail:userclass = {
    userId: '',
    username: '',
    krname: '',
    global_name: '',
    token: ''
}

function createPersistentStore(key:string, startValue:userclass) {
    const storedValue = localStorage.getItem(key);
    const initialValue = storedValue ? JSON.parse(storedValue) : startValue;
    const { subscribe, set, update } = writable(initialValue);
  
    return {
        subscribe,
        set: (value:any) => {
            localStorage.setItem(key, JSON.stringify(value));
            set(value);
        },
        update
    };
}

export const Logout = () => {
    useAuth.set(userDetail);
}

export const useAuth = createPersistentStore('useAuth', userDetail);