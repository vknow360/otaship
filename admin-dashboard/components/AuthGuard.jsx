import { useEffect, useState } from 'react';
import { useRouter } from 'next/router';

export default function AuthGuard({ children }) {
  const router = useRouter();
  const [authorized, setAuthorized] = useState(false);

  useEffect(() => {
    // Public paths
    if (router.pathname === '/login') {
      setAuthorized(true);
      return;
    }

    const token = localStorage.getItem('otaship_admin_token');
    if (!token) {
      setAuthorized(false);
      router.push({
        pathname: '/login',
        query: { returnUrl: router.asPath },
      });
    } else {
      setAuthorized(true);
    }
  }, [router]);

  // If on a protected route and not authorized, show nothing (or loading)
  if (!authorized && router.pathname !== '/login') {
    return null;
  }

  return children;
}
