SELECT id_reservation, id_staff, service_name, service_time, reservation_creator, 
       reservation_year, reservation_month, reservation_day, reservation_hour, 
       reservation_detail, reservation_verified, id_patient
  FROM reservation WHERE id_staff ='1635572582072481400' ORDER BY reservation_year DESC, reservation_month DESC,reservation_day DESC ;
