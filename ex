# Binarna postać liczby 
PROGRAM IS
	n, p
BEGIN
    READ n;
    FOR i FROM 0 DOWNTO 20 DO
	p:=n/2;
	p:=2*p;
	IF n>p THEN 
	    WRITE 1;
	ELSE 
	    WRITE 0;
	ENDIF
	n:=n/2;
    UNTIL n=0;
    ENDFOR
END
